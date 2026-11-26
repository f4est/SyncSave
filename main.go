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

	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	gdrive := storage.NewGDriveBackend(cfg)
	engine := syncer.NewEngine(cfg, gdrive, nil)
	engine.AddLog("info", "", "SyncSave запущен")

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
			w.SetSize(1120, 720, webview2.HintNone)
			w.Navigate(baseURL)
			w.Run()
		}()
	}

	trayIcon := tray.Start("SyncSave — Кооп-синхронизация сохранений", openWindow, func() {
		exitOnce.Do(func() { close(quitCh) })
	})
	defer trayIcon.Stop()

	if !cfg.Snapshot().StartMinimized {
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
				Backend            string `json:"backend"`
				LocalSyncDir       string `json:"local_sync_dir"`
				MinimizeToTray     bool   `json:"minimize_to_tray"`
				PollIntervalMs     int    `json:"poll_interval_ms"`
				GDriveClientID     string `json:"gdrive_client_id"`
				GDriveClientSecret string `json:"gdrive_client_secret"`
				GDriveFolderID     string `json:"gdrive_folder_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			err := cfg.Update(func(c *config.AppConfig) {
				if req.PlayerName != "" {
					c.PlayerName = req.PlayerName
				}
				if req.Backend == string(config.BackendLocal) {
					c.Backend = config.BackendLocal
				} else {
					c.Backend = config.BackendGDrive
				}
				if req.LocalSyncDir != "" {
					c.LocalSyncDir = req.LocalSyncDir
				}
				c.MinimizeToTray = req.MinimizeToTray
				if req.PollIntervalMs >= 2000 {
					c.PollIntervalMs = req.PollIntervalMs
				}
				c.GDrive.ClientID = req.GDriveClientID
				c.GDrive.ClientSecret = req.GDriveClientSecret
				c.GDrive.FolderID = req.GDriveFolderID
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
		if b.Name() == "local" {
			label = "Локальная папка: Готово"
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
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		idBytes := make([]byte, 6)
		_, _ = rand.Read(idBytes)
		g := config.GameProfile{
			ID:           hex.EncodeToString(idBytes),
			Name:         req.Name,
			LocalPath:    filepath.Clean(req.LocalPath),
			RemoteFolder: req.RemoteFolder,
			AutoSync:     req.AutoSync,
			KeepBackups:  5,
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
				AutoSync *bool `json:"auto_sync"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, err.Error())
				return
			}
			_ = cfg.Update(func(c *config.AppConfig) {
				for i := range c.Games {
					if c.Games[i].ID == id && req.AutoSync != nil {
						c.Games[i].AutoSync = *req.AutoSync
					}
				}
			})
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})

		case action == "" && r.Method == http.MethodDelete:
			_ = cfg.Update(func(c *config.AppConfig) {
				filtered := make([]config.GameProfile, 0, len(c.Games))
				for _, g := range c.Games {
					if g.ID != id {
						filtered = append(filtered, g)
					}
				}
				c.Games = filtered
			})
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
