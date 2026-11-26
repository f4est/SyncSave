package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jchv/go-webview2"

	"syncsave/internal/config"
	"syncsave/internal/presets"
	"syncsave/internal/storage"
	"syncsave/internal/syncer"
	"syncsave/internal/tray"
	"syncsave/ui"
)

func main() {
	releaseMutex, ok := tray.AcquireSingleInstance("Global\\SyncSaveSingleInstanceMutex")
	if !ok {
		return
	}
	defer releaseMutex()

	cfg, loadErr := config.Load()
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	// Sync autostart state with actual Windows Registry
	if realAuto := config.IsAutostartEnabled(); realAuto != cfg.Snapshot().Autostart {
		_ = cfg.Update(func(c *config.AppConfig) { c.Autostart = realAuto })
	}

	startMinFlag := false
	for _, arg := range os.Args[1:] {
		if arg == "--minimized" {
			startMinFlag = true
		}
	}

	gdrive := storage.NewGDriveBackend(cfg)
	engine := syncer.NewEngine(cfg, gdrive, nil)
	if loadErr != nil {
		engine.AddLog("warn", "", loadErr.Error())
	} else {
		engine.AddLog("info", "", "SyncSave запущен")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start low-overhead background file watcher
	go engine.RunWatcher(ctx)

	// Start local HTTP API + embedded UI server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}
	port := ln.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	mux := buildMux(cfg, gdrive, engine)
	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()
	defer func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutCancel()
		_ = srv.Shutdown(shutCtx)
	}()

	var (
		trayIcon *tray.Tray
		winMu    sync.Mutex
		winOpen  bool
		exitOnce sync.Once
		quitCh   = make(chan struct{})
	)

	openWindow := func() {
		winMu.Lock()
		if winOpen {
			winMu.Unlock()
			return
		}
		winOpen = true
		winMu.Unlock()

		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			defer func() {
				winMu.Lock()
				winOpen = false
				winMu.Unlock()
				if !cfg.Snapshot().MinimizeToTray {
					exitOnce.Do(func() { close(quitCh) })
				} else if cfg.Snapshot().MemorySaver {
					go tray.TrimMemory()
				}
			}()

			w := webview2.NewWithOptions(webview2.WebViewOptions{
				Debug:     false,
				AutoFocus: true,
				DataPath:  filepath.Join(config.AppDataDir(), "webview2"),
				WindowOptions: webview2.WindowOptions{
					Title:  "SyncSave — Co-op Save Sync",
					Width:  1120,
					Height: 720,
					Center: true,
				},
			})
			if w == nil {
				// Fallback to default browser if WebView2 fails to initialize
				_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", baseURL).Start()
				return
			}
			defer w.Destroy()
			trayIcon.ApplyWindowIcon(w.Window())
			w.SetSize(1120, 720, webview2.HintNone)
			w.Navigate(baseURL)
			w.Run()
		}()
	}

	trayIcon = tray.Start("SyncSave — Кооп-синхронизация сохранений", tray.Callbacks{
		OnShow: openWindow,
		OnSyncAll: func() {
			go func() {
				sCtx, sCancel := context.WithTimeout(ctx, 5*time.Minute)
				defer sCancel()
				engine.SyncAll(sCtx)
			}()
		},
		OnTogglePause: func() {
			_ = cfg.Update(func(c *config.AppConfig) {
				c.AutoSyncPaused = !c.AutoSyncPaused
			})
			if cfg.Snapshot().AutoSyncPaused {
				engine.AddLog("warn", "", "Авто-синхронизация поставлена на паузу")
			} else {
				engine.AddLog("info", "", "Авто-синхронизация возобновлена")
			}
		},
		OnOpenBackups: func() {
			_ = os.MkdirAll(config.BackupsDir(), 0o755)
			_ = exec.Command("explorer", config.BackupsDir()).Start()
		},
		OnToggleStartup: func() {
			snap := cfg.Snapshot()
			next := !snap.Autostart
			if err := config.SetAutostart(next, snap.StartMinimized); err == nil {
				_ = cfg.Update(func(c *config.AppConfig) {
					c.Autostart = next
				})
			}
		},
		OnExit: func() {
			exitOnce.Do(func() { close(quitCh) })
		},
		IsSingleClick: func() bool { return cfg.Snapshot().TraySingleClick },
		IsPaused:      func() bool { return cfg.Snapshot().AutoSyncPaused },
		IsAutostart:   func() bool { return cfg.Snapshot().Autostart },
	})
	defer trayIcon.Stop()
	engine.SetNotifier(trayIcon.ShowBalloon)

	if cfg.Snapshot().AutoPullOnStart {
		go func() {
			sCtx, sCancel := context.WithTimeout(ctx, 3*time.Minute)
			defer sCancel()
			engine.SyncAll(sCtx)
		}()
	}

	if !startMinFlag && !cfg.Snapshot().StartMinimized {
		openWindow()
	}

	<-quitCh
}

func buildMux(cfg *config.AppConfig, gdrive *storage.GDriveBackend, engine *syncer.Engine) *http.ServeMux {
	mux := http.NewServeMux()

	staticFS, _ := fs.Sub(ui.Static, "static")
	fileServer := http.FileServer(http.FS(staticFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			snap := cfg.Snapshot()
			// Mask secret slightly? Keep it editable in local app.
			writeJSON(w, http.StatusOK, snap)
		case http.MethodPost:
			var req struct {
				PlayerName         string `json:"player_name"`
				Language           string `json:"language"`
				Theme              string `json:"theme"`
				Backend            string `json:"backend"`
				LocalSyncDir       string `json:"local_sync_dir"`
				MinimizeToTray     bool   `json:"minimize_to_tray"`
				StartMinimized     bool   `json:"start_minimized"`
				Autostart          bool   `json:"autostart"`
				TraySingleClick    bool   `json:"tray_single_click"`
				TrayNotifications  bool   `json:"tray_notifications"`
				MemorySaver        bool   `json:"memory_saver"`
				AutoSyncPaused     bool   `json:"auto_sync_paused"`
				AutoPullOnStart    bool   `json:"auto_pull_on_start"`
				DebounceSeconds    int    `json:"debounce_seconds"`
				DefaultKeepBackups int    `json:"default_keep_backups"`
				PollIntervalMs     int    `json:"poll_interval_ms"`
				GDriveClientID     string `json:"gdrive_client_id"`
				GDriveClientSecret string `json:"gdrive_client_secret"`
				GDriveFolderID     string `json:"gdrive_folder_id"`
				WebDAVURL          string `json:"webdav_url"`
				WebDAVUsername     string `json:"webdav_username"`
				WebDAVPassword     string `json:"webdav_password"`
				WebDAVBasePath     string `json:"webdav_base_path"`
				CustomPullCmd      string `json:"custom_pull_cmd"`
				CustomPushCmd      string `json:"custom_push_cmd"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			if err := config.SetAutostart(req.Autostart, req.StartMinimized); err != nil {
				writeErr(w, http.StatusInternalServerError, "ошибка автозапуска: "+err.Error())
				return
			}
			err := cfg.Update(func(c *config.AppConfig) {
				if req.PlayerName != "" {
					c.PlayerName = req.PlayerName
				}
				if req.Language != "" {
					c.Language = req.Language
				}
				if req.Theme != "" {
					c.Theme = req.Theme
				}
				switch req.Backend {
				case string(config.BackendLocal):
					c.Backend = config.BackendLocal
				case string(config.BackendWebDAV):
					c.Backend = config.BackendWebDAV
				case string(config.BackendCustom):
					c.Backend = config.BackendCustom
				default:
					c.Backend = config.BackendGDrive
				}
				if req.LocalSyncDir != "" {
					c.LocalSyncDir = req.LocalSyncDir
				}
				c.MinimizeToTray = req.MinimizeToTray
				c.StartMinimized = req.StartMinimized
				c.Autostart = req.Autostart
				c.TraySingleClick = req.TraySingleClick
				c.TrayNotifications = req.TrayNotifications
				c.MemorySaver = req.MemorySaver
				c.AutoSyncPaused = req.AutoSyncPaused
				c.AutoPullOnStart = req.AutoPullOnStart
				if req.DebounceSeconds >= 2 {
					c.DebounceSeconds = req.DebounceSeconds
				}
				if req.DefaultKeepBackups >= 1 {
					c.DefaultKeepBackups = req.DefaultKeepBackups
				}
				if req.PollIntervalMs >= 2000 {
					c.PollIntervalMs = req.PollIntervalMs
				}
				c.GDrive.ClientID = req.GDriveClientID
				c.GDrive.ClientSecret = req.GDriveClientSecret
				c.GDrive.FolderID = req.GDriveFolderID
				c.WebDAV.URL = req.WebDAVURL
				c.WebDAV.Username = req.WebDAVUsername
				c.WebDAV.Password = req.WebDAVPassword
				c.WebDAV.BasePath = req.WebDAVBasePath
				c.CustomSync.PullCmd = req.CustomPullCmd
				c.CustomSync.PushCmd = req.CustomPushCmd
			})
			if err != nil {
				writeErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, cfg.Snapshot())
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/presets", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, presets.BuiltinPresets())
	})

	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, engine.Logs())
	})

	mux.HandleFunc("/api/backend/check", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		b := engine.Backend()
		if err := b.CheckConnection(ctx); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		label := "Google Drive: Подключено"
		switch b.Name() {
		case "local":
			label = "Локальная папка: Готово"
		case "webdav":
			label = "WebDAV: Подключено"
		case "custom":
			label = "Своя команда: Готово"
		}
		writeJSON(w, http.StatusOK, map[string]string{"label": label})
	})

	mux.HandleFunc("/api/remote-worlds", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		worlds, err := engine.Backend().ListRemoteWorlds(ctx)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, worlds)
	})

	mux.HandleFunc("/api/gdrive/oauth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		email, err := gdrive.StartOAuthFlow(r.Context(), req.ClientID, req.ClientSecret)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		engine.AddLog("success", "", "Google Drive авторизован ("+email+")")
		writeJSON(w, http.StatusOK, map[string]string{"email": email})
	})

	mux.HandleFunc("/api/games", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Name         string `json:"name"`
			LocalPath    string `json:"local_path"`
			RemoteFolder string `json:"remote_folder"`
			AutoSync     bool   `json:"auto_sync"`
			KeepBackups  int    `json:"keep_backups"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.LocalPath = strings.TrimSpace(req.LocalPath)
		req.RemoteFolder = strings.TrimSpace(req.RemoteFolder)
		if req.Name == "" || req.LocalPath == "" || req.RemoteFolder == "" {
			writeErr(w, http.StatusBadRequest, "название, локальный путь и имя мира обязательны")
			return
		}
		if strings.ContainsAny(req.RemoteFolder, `/\`) || strings.Contains(req.RemoteFolder, "..") {
			writeErr(w, http.StatusBadRequest, "имя общего мира не должно содержать / \\ или ..")
			return
		}
		keep := req.KeepBackups
		if keep <= 0 {
			keep = 5
		}
		idBytes := make([]byte, 6)
		_, _ = rand.Read(idBytes)
		g := config.GameProfile{
			ID:           hex.EncodeToString(idBytes),
			Name:         req.Name,
			LocalPath:    filepath.Clean(req.LocalPath),
			RemoteFolder: req.RemoteFolder,
			AutoSync:     req.AutoSync,
			KeepBackups:  keep,
		}
		if err := cfg.Update(func(c *config.AppConfig) {
			c.Games = append(c.Games, g)
		}); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		engine.AddLog("info", g.Name, "Игра добавлена в SyncSave")
		writeJSON(w, http.StatusOK, g)
	})

	mux.HandleFunc("/api/games/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/games/")
		parts := strings.Split(rest, "/")
		id := parts[0]
		action := ""
		if len(parts) > 1 {
			action = parts[1]
		}

		switch {
		case action == "" && r.Method == http.MethodPatch:
			var req struct {
				AutoSync    *bool `json:"auto_sync"`
				KeepBackups *int  `json:"keep_backups"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			if err := cfg.Update(func(c *config.AppConfig) {
				for i := range c.Games {
					if c.Games[i].ID == id {
						if req.AutoSync != nil {
							c.Games[i].AutoSync = *req.AutoSync
						}
						if req.KeepBackups != nil && *req.KeepBackups > 0 {
							c.Games[i].KeepBackups = *req.KeepBackups
						}
					}
				}
			}); err != nil {
				writeErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})

		case action == "" && r.Method == http.MethodDelete:
			if err := cfg.Update(func(c *config.AppConfig) {
				filtered := make([]config.GameProfile, 0, len(c.Games))
				for _, g := range c.Games {
					if g.ID != id {
						filtered = append(filtered, g)
					}
				}
				c.Games = filtered
			}); err != nil {
				writeErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})

		case action == "status" && r.Method == http.MethodGet:
			ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
			defer cancel()
			st := engine.GetGameStatus(ctx, id)
			writeJSON(w, http.StatusOK, st)

		case action == "push" && r.Method == http.MethodPost:
			var req struct {
				Note  string `json:"note"`
				Force bool   `json:"force"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
			defer cancel()
			ver, err := engine.Push(ctx, id, req.Note, req.Force)
			if err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, ver)

		case action == "pull" && r.Method == http.MethodPost:
			var req struct {
				Version int  `json:"version"`
				Force   bool `json:"force"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
			defer cancel()
			ver, err := engine.Pull(ctx, id, req.Version, req.Force)
			if err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, ver)

		case action == "lock" && r.Method == http.MethodPost:
			var req struct {
				Acquire bool `json:"acquire"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
			defer cancel()
			if err := engine.SetHostLock(ctx, id, req.Acquire); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})

		default:
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/api/open-folder", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Path string `json:"path"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Path != "" {
			_ = exec.Command("explorer", filepath.Clean(req.Path)).Start()
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("/api/browse-folder", func(w http.ResponseWriter, r *http.Request) {
		// Lightweight native Windows FolderBrowserDialog via PowerShell
		ps := `Add-Type -AssemblyName System.Windows.Forms; $d = New-Object System.Windows.Forms.FolderBrowserDialog; $d.Description = 'Выберите папку сохранений'; if ($d.ShowDialog() -eq 'OK') { Write-Output $d.SelectedPath }`
		out, err := exec.Command("powershell", "-NoProfile", "-STA", "-Command", ps).Output()
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]string{"path": ""})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"path": strings.TrimSpace(string(out))})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
