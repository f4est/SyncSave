package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"syncsave/internal/archive"
	"syncsave/internal/config"
	"syncsave/internal/storage"
)

// SyncState describes the relationship between local saves and the shared cloud world.
type SyncState string

const (
	StateSynced       SyncState = "synced"
	StateLocalAhead   SyncState = "local_ahead"
	StateRemoteAhead  SyncState = "remote_ahead"
	StateConflict     SyncState = "conflict"
	StateNotSynced    SyncState = "not_synced"
	StateMissingLocal SyncState = "missing_local"
	StateError        SyncState = "error"
)

// GameStatus is returned to the UI for each configured game.
type GameStatus struct {
	Game          config.GameProfile     `json:"game"`
	State         SyncState              `json:"state"`
	LocalHash     string                 `json:"local_hash"`
	LocalSize     int64                  `json:"local_size"`
	LocalFiles    int                    `json:"local_files"`
	RemoteVersion int                    `json:"remote_version"`
	RemoteAuthor  string                 `json:"remote_author,omitempty"`
	RemoteUpdated string                 `json:"remote_updated,omitempty"`
	HostLock      *storage.HostLock      `json:"host_lock,omitempty"`
	Versions      []storage.SaveVersion  `json:"versions"`
	Events        []storage.WorldEvent   `json:"events,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

// ActivityEntry represents one event in the live SyncSave log.
type ActivityEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"` // info, success, warn, error
	Game    string `json:"game,omitempty"`
	Message string `json:"message"`
}

// Engine orchestrates co-op save syncing, conflict prevention, backups, and background watching.
type Engine struct {
	cfg      *config.AppConfig
	gdrive   *storage.GDriveBackend
	mu       sync.Mutex
	busy     map[string]bool
	logs     []ActivityEntry
	logMu    sync.RWMutex
	onEvent  func()
	onNotify func(title, msg string)

	// Lightweight stat cache for 0% CPU idle watching
	watchMu          sync.Mutex
	lastStatSig      map[string]string
	pendingChange    map[string]time.Time
	lastRemoteCheck  time.Time
	notifiedConflict map[string]int
}

func NewEngine(cfg *config.AppConfig, gdrive *storage.GDriveBackend, onEvent func()) *Engine {
	return &Engine{
		cfg:              cfg,
		gdrive:           gdrive,
		busy:             make(map[string]bool),
		logs:             make([]ActivityEntry, 0, 100),
		onEvent:          onEvent,
		lastStatSig:      make(map[string]string),
		pendingChange:    make(map[string]time.Time),
		notifiedConflict: make(map[string]int),
	}
}

// SetNotifier sets an optional callback for desktop/tray notifications.
func (e *Engine) SetNotifier(fn func(title, msg string)) {
	e.onNotify = fn
}

func (e *Engine) notify(title, msg string) {
	if e.onNotify != nil && e.cfg.Snapshot().TrayNotifications {
		e.onNotify(title, msg)
	}
}

func (e *Engine) Backend() storage.Backend {
	snap := e.cfg.Snapshot()
	switch snap.Backend {
	case config.BackendLocal:
		return storage.NewLocalBackend(snap.LocalSyncDir)
	case config.BackendWebDAV:
		return storage.NewWebDAVBackend(snap.WebDAV.URL, snap.WebDAV.Username, snap.WebDAV.Password, snap.WebDAV.BasePath)
	case config.BackendS3:
		return storage.NewS3Backend(snap.S3.Endpoint, snap.S3.Region, snap.S3.Bucket, snap.S3.AccessKey, snap.S3.SecretKey, snap.S3.Prefix)
	case config.BackendGit:
		return storage.NewGitBackend(snap.Git.RepoURL, snap.Git.Branch, filepath.Join(config.AppDataDir(), "GitSyncRepo"))
	case config.BackendLAN:
		if snap.LAN.HostEnabled && snap.LAN.PeerURL == "" {
			return storage.NewLocalBackend(snap.LocalSyncDir)
		}
		return storage.NewLANBackend(snap.LAN.PeerURL, snap.LAN.RoomPIN)
	case config.BackendCustom:
		return storage.NewCustomCmdBackend(snap.LocalSyncDir, snap.CustomSync.PullCmd, snap.CustomSync.PushCmd)
	default:
		return e.gdrive
	}
}

func (e *Engine) AddLog(level, game, msg string) {
	e.logMu.Lock()
	entry := ActivityEntry{
		Time:    time.Now().Format("15:04:05"),
		Level:   level,
		Game:    game,
		Message: msg,
	}
	e.logs = append([]ActivityEntry{entry}, e.logs...)
	if len(e.logs) > 100 {
		e.logs = e.logs[:100]
	}
	e.logMu.Unlock()
	if e.onEvent != nil {
		e.onEvent()
	}
}

func (e *Engine) Logs() []ActivityEntry {
	e.logMu.RLock()
	defer e.logMu.RUnlock()
	cp := make([]ActivityEntry, len(e.logs))
	copy(cp, e.logs)
	return cp
}

func (e *Engine) lockGame(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.busy[id] {
		return false
	}
	e.busy[id] = true
	return true
}

func (e *Engine) unlockGame(id string) {
	e.mu.Lock()
	delete(e.busy, id)
	e.mu.Unlock()
}

func (e *Engine) findGame(id string) (config.GameProfile, bool) {
	snap := e.cfg.Snapshot()
	for _, g := range snap.Games {
		if g.ID == id {
			return g, true
		}
	}
	return config.GameProfile{}, false
}

// GetGameStatus inspects local files and remote manifest to determine sync state.
func (e *Engine) GetGameStatus(ctx context.Context, id string) GameStatus {
	g, ok := e.findGame(id)
	if !ok {
		return GameStatus{State: StateError, Error: "игра не найдена"}
	}
	st := GameStatus{Game: g, Versions: []storage.SaveVersion{}}

	localHash, localSize, localFiles, localErr := archive.DirectoryHash(g.LocalPath)
	if localErr == nil {
		st.LocalHash = localHash
		st.LocalSize = localSize
		st.LocalFiles = localFiles
	}

	b := e.Backend()
	manifest, err := b.GetManifest(ctx, g.RemoteFolder)
	if err != nil {
		if localErr != nil {
			st.State = StateMissingLocal
		} else {
			st.State = StateError
		}
		st.Error = err.Error()
		return st
	}

	if manifest != nil {
		st.RemoteVersion = manifest.LatestVersion
		st.RemoteAuthor = manifest.UpdatedBy
		st.RemoteUpdated = manifest.UpdatedAt
		st.HostLock = manifest.ActiveHostLock
		st.Versions = manifest.Versions
		st.Events = manifest.Events
	}

	if localErr != nil {
		st.State = StateMissingLocal
		return st
	}

	localDirty := g.LastLocalHash == "" || localHash != g.LastLocalHash
	if manifest == nil || manifest.LatestVersion == 0 {
		if localFiles > 0 {
			st.State = StateLocalAhead
		} else {
			st.State = StateNotSynced
		}
		return st
	}

	remoteAhead := manifest.LatestVersion > g.LastSyncedVer
	// Check if local hash already matches the latest remote version's dir hash
	if len(manifest.Versions) > 0 {
		latest := manifest.Versions[0]
		if latest.DirHash == localHash {
			st.State = StateSynced
			return st
		}
	}

	switch {
	case !localDirty && !remoteAhead:
		st.State = StateSynced
	case localDirty && !remoteAhead:
		st.State = StateLocalAhead
	case !localDirty && remoteAhead:
		st.State = StateRemoteAhead
	case localDirty && remoteAhead:
		st.State = StateConflict
	}
	return st
}

// Push packs and uploads the local save folder as a new version in the shared world.
func (e *Engine) Push(ctx context.Context, id, note string, force bool) (*storage.SaveVersion, error) {
	if !e.lockGame(id) {
		return nil, fmt.Errorf("синхронизация для этой игры уже выполняется")
	}
	defer e.unlockGame(id)

	g, ok := e.findGame(id)
	if !ok {
		return nil, fmt.Errorf("игра не найдена")
	}
	snap := e.cfg.Snapshot()
	device, _ := os.Hostname()

	dirHash, rawSize, fileCount, err := archive.DirectoryHash(g.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения папки сохранений: %w", err)
	}
	if fileCount == 0 {
		return nil, fmt.Errorf("папка сохранений пуста")
	}

	b := e.Backend()
	manifest, err := b.GetManifest(ctx, g.RemoteFolder)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить manifest: %w", err)
	}
	if manifest == nil {
		manifest = &storage.WorldManifest{
			GameName: g.Name,
			Versions: []storage.SaveVersion{},
		}
	}

	// Conflict prevention
	if !force && manifest.LatestVersion > g.LastSyncedVer {
		return nil, fmt.Errorf("конфликт версий: в облаке уже есть версия v%d (от %s), а у вас v%d. Сначала скачайте обновление или нажмите «Принудительно»",
			manifest.LatestVersion, manifest.UpdatedBy, g.LastSyncedVer)
	}
	if !force && manifest.ActiveHostLock != nil && manifest.ActiveHostLock.Holder != snap.PlayerName {
		return nil, fmt.Errorf("сейчас хостит игрок %s (с %s). Попросите его освободить хост или загрузите принудительно",
			manifest.ActiveHostLock.Holder, manifest.ActiveHostLock.AcquiredAt)
	}

	if len(manifest.Versions) > 0 && manifest.Versions[0].DirHash == dirHash && !force {
		return nil, fmt.Errorf("изменений нет — локальное сохранение совпадает с последней версией v%d", manifest.LatestVersion)
	}

	nextVer := manifest.LatestVersion + 1
	ts := time.Now().Format("20060102_150405")
	zipName := fmt.Sprintf("v%04d_%s_%s.zip", nextVer, sanitizeName(snap.PlayerName), ts)
	tmpZip := filepath.Join(os.TempDir(), "syncsave_"+zipName)
	defer os.Remove(tmpZip)

	e.AddLog("info", g.Name, fmt.Sprintf("Упаковка сохранения v%d...", nextVer))
	zipSize, zipSHA, err := archive.PackDirectory(g.LocalPath, tmpZip)
	if err != nil {
		e.AddLog("error", g.Name, "Ошибка упаковки: "+err.Error())
		return nil, err
	}

	e.AddLog("info", g.Name, fmt.Sprintf("Загрузка v%d (%.2f МБ)...", nextVer, float64(zipSize)/(1024*1024)))
	remoteID, err := b.UploadArchive(ctx, g.RemoteFolder, zipName, tmpZip)
	if err != nil {
		e.AddLog("error", g.Name, "Ошибка загрузки: "+err.Error())
		return nil, err
	}

	nowStr := time.Now().Format("2006-01-02 15:04:05")
	ver := storage.SaveVersion{
		Version:    nextVer,
		Author:     snap.PlayerName,
		Device:     device,
		Timestamp:  nowStr,
		Note:       strings.TrimSpace(note),
		DirHash:    dirHash,
		ArchiveSHA: zipSHA,
		RawSize:    rawSize,
		ZipSize:    zipSize,
		FileCount:  fileCount,
		FileName:   zipName,
		RemoteID:   remoteID,
	}

	manifest.LatestVersion = nextVer
	manifest.UpdatedAt = nowStr
	manifest.UpdatedBy = snap.PlayerName
	manifest.Versions = append([]storage.SaveVersion{ver}, manifest.Versions...)
	if len(manifest.Versions) > 50 {
		manifest.Versions = manifest.Versions[:50]
	}
	appendWorldEvent(manifest, snap.PlayerName, "success", fmt.Sprintf("Загрузил(а) версию v%d%s", nextVer, noteSuffix(note)))

	if err := b.PutManifest(ctx, g.RemoteFolder, manifest); err != nil {
		e.AddLog("error", g.Name, "Ошибка обновления manifest: "+err.Error())
		return nil, err
	}

	_ = e.cfg.Update(func(c *config.AppConfig) {
		for i := range c.Games {
			if c.Games[i].ID == id {
				c.Games[i].LastSyncedVer = nextVer
				c.Games[i].LastLocalHash = dirHash
				c.Games[i].LastSyncAt = nowStr
				break
			}
		}
	})

	e.AddLog("success", g.Name, fmt.Sprintf("Загружена версия v%d (%s)", nextVer, snap.PlayerName))
	e.notify("SyncSave — "+g.Name, fmt.Sprintf("Сохранение v%d загружено в облако", nextVer))
	return &ver, nil
}

// Pull downloads the latest (or specified) version, creates a local safety backup, and unpacks it.
func (e *Engine) Pull(ctx context.Context, id string, targetVersion int, force bool) (*storage.SaveVersion, error) {
	if !e.lockGame(id) {
		return nil, fmt.Errorf("синхронизация для этой игры уже выполняется")
	}
	defer e.unlockGame(id)

	g, ok := e.findGame(id)
	if !ok {
		return nil, fmt.Errorf("игра не найдена")
	}

	b := e.Backend()
	manifest, err := b.GetManifest(ctx, g.RemoteFolder)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить manifest: %w", err)
	}
	if manifest == nil || len(manifest.Versions) == 0 {
		return nil, fmt.Errorf("в облаке пока нет версий для этого мира")
	}

	var chosen *storage.SaveVersion
	if targetVersion <= 0 {
		chosen = &manifest.Versions[0]
	} else {
		for i := range manifest.Versions {
			if manifest.Versions[i].Version == targetVersion {
				chosen = &manifest.Versions[i]
				break
			}
		}
	}
	if chosen == nil {
		return nil, fmt.Errorf("версия v%d не найдена", targetVersion)
	}

	// Check local dirty state before overwriting
	localHash, _, localFiles, localErr := archive.DirectoryHash(g.LocalPath)
	if localErr == nil && localFiles > 0 && g.LastLocalHash != "" && localHash != g.LastLocalHash && !force {
		return nil, fmt.Errorf("локальные сохранения были изменены. Нажмите «Принудительно скачать» (старая версия сохранится в локальный бэкап)")
	}

	// Create local safety backup before replacing
	if localErr == nil && localFiles > 0 {
		e.createLocalBackup(g)
	}

	tmpZip := filepath.Join(os.TempDir(), "syncsave_pull_"+chosen.FileName)
	defer os.Remove(tmpZip)

	e.AddLog("info", g.Name, fmt.Sprintf("Скачивание v%d (%s)...", chosen.Version, chosen.Author))
	if err := b.DownloadArchive(ctx, g.RemoteFolder, *chosen, tmpZip); err != nil {
		e.AddLog("error", g.Name, "Ошибка скачивания: "+err.Error())
		return nil, err
	}

	if chosen.ArchiveSHA != "" {
		gotSHA, err := archive.FileSHA256(tmpZip)
		if err != nil || gotSHA != chosen.ArchiveSHA {
			e.AddLog("error", g.Name, "Ошибка проверки SHA-256 архива")
			return nil, fmt.Errorf("контрольная сумма архива не совпадает (файл повреждён)")
		}
	}

	if err := archive.UnpackDirectory(tmpZip, g.LocalPath); err != nil {
		e.AddLog("error", g.Name, "Ошибка распаковки: "+err.Error())
		return nil, err
	}

	nowStr := time.Now().Format("2006-01-02 15:04:05")
	newHash, _, _, _ := archive.DirectoryHash(g.LocalPath)
	_ = e.cfg.Update(func(c *config.AppConfig) {
		for i := range c.Games {
			if c.Games[i].ID == id {
				c.Games[i].LastSyncedVer = chosen.Version
				c.Games[i].LastLocalHash = newHash
				c.Games[i].LastSyncAt = nowStr
				break
			}
		}
	})

	e.AddLog("success", g.Name, fmt.Sprintf("Установлена версия v%d (от %s)", chosen.Version, chosen.Author))
	e.notify("SyncSave — "+g.Name, fmt.Sprintf("Установлена версия v%d (от %s)", chosen.Version, chosen.Author))

	// Record pull event in shared world manifest (best-effort)
	player := e.cfg.Snapshot().PlayerName
	appendWorldEvent(manifest, player, "info", fmt.Sprintf("Скачал(а) версию v%d", chosen.Version))
	_ = b.PutManifest(ctx, g.RemoteFolder, manifest)
	return chosen, nil
}

// SetHostLock acquires or releases the co-op host lock in manifest.json.
func (e *Engine) SetHostLock(ctx context.Context, id string, acquire bool) error {
	g, ok := e.findGame(id)
	if !ok {
		return fmt.Errorf("игра не найдена")
	}
	snap := e.cfg.Snapshot()
	device, _ := os.Hostname()
	b := e.Backend()

	manifest, err := b.GetManifest(ctx, g.RemoteFolder)
	if err != nil {
		return err
	}
	if manifest == nil {
		manifest = &storage.WorldManifest{GameName: g.Name, Versions: []storage.SaveVersion{}}
	}

	if acquire {
		if manifest.ActiveHostLock != nil && manifest.ActiveHostLock.Holder != snap.PlayerName {
			return fmt.Errorf("хост уже занят игроком %s", manifest.ActiveHostLock.Holder)
		}
		manifest.ActiveHostLock = &storage.HostLock{
			Holder:     snap.PlayerName,
			Device:     device,
			AcquiredAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		e.AddLog("info", g.Name, fmt.Sprintf("%s взял(а) роль хоста", snap.PlayerName))
		appendWorldEvent(manifest, snap.PlayerName, "info", "Взял(а) роль хоста")
	} else {
		manifest.ActiveHostLock = nil
		e.AddLog("info", g.Name, fmt.Sprintf("%s освободил(а) роль хоста", snap.PlayerName))
		appendWorldEvent(manifest, snap.PlayerName, "info", "Освободил(а) роль хоста")
	}
	return b.PutManifest(ctx, g.RemoteFolder, manifest)
}

func appendWorldEvent(m *storage.WorldManifest, player, level, msg string) {
	ev := storage.WorldEvent{
		Time:    time.Now().Format("02.01 15:04:05"),
		Player:  player,
		Level:   level,
		Message: msg,
	}
	m.Events = append([]storage.WorldEvent{ev}, m.Events...)
	if len(m.Events) > 60 {
		m.Events = m.Events[:60]
	}
}

func noteSuffix(note string) string {
	n := strings.TrimSpace(note)
	if n == "" {
		return ""
	}
	return " (" + n + ")"
}

func (e *Engine) createLocalBackup(g config.GameProfile) {
	keep := g.KeepBackups
	if keep <= 0 {
		keep = e.cfg.Snapshot().DefaultKeepBackups
	}
	if keep <= 0 {
		keep = 5
	}
	backupDir := filepath.Join(config.BackupsDir(), sanitizeName(g.ID))
	ts := time.Now().Format("20060102_150405")
	dst := filepath.Join(backupDir, fmt.Sprintf("backup_%s.zip", ts))
	if _, _, err := archive.PackDirectory(g.LocalPath, dst); err == nil {
		e.AddLog("info", g.Name, "Создан локальный бэкап перед заменой")
	}

	// Prune old backups
	entries, err := os.ReadDir(backupDir)
	if err != nil || len(entries) <= keep {
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for i := 0; i < len(entries)-keep; i++ {
		_ = os.Remove(filepath.Join(backupDir, entries[i].Name()))
	}
}

// RunWatcher starts the ultra-lightweight background auto-sync loop.
// It uses fast file stat signatures (mtime + size) and only hashes/uploads after an 8s quiet debounce.
func (e *Engine) RunWatcher(ctx context.Context) {
	for {
		interval := time.Duration(e.cfg.Snapshot().PollIntervalMs) * time.Millisecond
		if interval < 2*time.Second {
			interval = 5 * time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			e.watchTick(ctx)
		}
	}
}

// SyncAll checks all configured games and pushes local changes or pulls updates if safe.
func (e *Engine) SyncAll(ctx context.Context) {
	snap := e.cfg.Snapshot()
	if len(snap.Games) == 0 {
		return
	}
	e.AddLog("info", "", "Запуск синхронизации всех игр...")
	for _, g := range snap.Games {
		st := e.GetGameStatus(ctx, g.ID)
		switch st.State {
		case StateLocalAhead:
			_, _ = e.Push(ctx, g.ID, "Sync All", false)
		case StateRemoteAhead:
			_, _ = e.Pull(ctx, g.ID, 0, false)
		case StateConflict:
			e.AddLog("warn", g.Name, "Пропущено при Sync All: конфликт версий")
			e.notify("SyncSave — Конфликт", g.Name+": требуется ручное решение конфликта")
		}
	}
}

func (e *Engine) watchTick(ctx context.Context) {
	snap := e.cfg.Snapshot()
	if snap.AutoSyncPaused {
		return
	}
	now := time.Now()
	debounce := time.Duration(snap.DebounceSeconds) * time.Second
	if debounce < 2*time.Second {
		debounce = 8 * time.Second
	}

	for _, g := range snap.Games {
		if !g.AutoSync {
			continue
		}
		sig := quickDirStatSig(g.LocalPath)
		if sig == "" {
			continue
		}

		e.watchMu.Lock()
		prev, hasPrev := e.lastStatSig[g.ID]
		if !hasPrev {
			e.lastStatSig[g.ID] = sig
			e.watchMu.Unlock()
			continue
		}
		if sig != prev {
			e.lastStatSig[g.ID] = sig
			e.pendingChange[g.ID] = now
			e.watchMu.Unlock()
			continue
		}
		changedAt, pending := e.pendingChange[g.ID]
		if pending && now.Sub(changedAt) >= debounce {
			delete(e.pendingChange, g.ID)
			e.watchMu.Unlock()

			// Debounce settled — auto-push
			go func(gameID, gameName string) {
				pushCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
				defer cancel()
				if _, err := e.Push(pushCtx, gameID, "Auto-sync", false); err != nil {
					if !strings.Contains(err.Error(), "изменений нет") {
						e.AddLog("warn", gameName, "Auto-sync: "+err.Error())
					}
				}
			}(g.ID, g.Name)
		} else {
			e.watchMu.Unlock()
		}
	}

	if snap.AutoPullEnabled {
		remoteInterval := time.Duration(snap.RemotePollSeconds) * time.Second
		if remoteInterval < 10*time.Second {
			remoteInterval = 60 * time.Second
		}
		if now.Sub(e.lastRemoteCheck) >= remoteInterval {
			e.lastRemoteCheck = now
			go e.remotePullTick(ctx)
		}
	}
}

func (e *Engine) remotePullTick(ctx context.Context) {
	snap := e.cfg.Snapshot()
	if snap.AutoSyncPaused || !snap.AutoPullEnabled {
		return
	}
	for _, g := range snap.Games {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		st := e.GetGameStatus(checkCtx, g.ID)
		switch st.State {
		case StateRemoteAhead:
			if _, err := e.Pull(checkCtx, g.ID, 0, false); err != nil {
				e.AddLog("warn", g.Name, "Auto-pull: "+err.Error())
			}
		case StateConflict:
			if e.notifiedConflict[g.ID] != st.RemoteVersion {
				e.notifiedConflict[g.ID] = st.RemoteVersion
				e.AddLog("warn", g.Name, fmt.Sprintf("Конфликт с облачной версией v%d", st.RemoteVersion))
				e.notify("SyncSave — Конфликт", fmt.Sprintf("%s: новая версия v%d конфликтует с локальными изменениями", g.Name, st.RemoteVersion))
			}
		}
		cancel()
	}
}

func quickDirStatSig(dir string) string {
	var totalSize int64
	var maxMod int64
	var count int
	err := filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".lock") {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return nil
		}
		count++
		totalSize += fi.Size()
		if m := fi.ModTime().UnixNano(); m > maxMod {
			maxMod = m
		}
		return nil
	})
	if err != nil || count == 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d:%d", count, totalSize, maxMod)
}

func sanitizeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "player"
	}
	return b.String()
}
