package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// BackendType defines how saves are synchronized online.
type BackendType string

const (
	BackendGDrive BackendType = "gdrive"
	BackendLocal  BackendType = "local"
)

// GameProfile represents a configured co-op game and its shared world.
type GameProfile struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	PresetID      string `json:"preset_id,omitempty"`
	LocalPath     string `json:"local_path"`
	RemoteFolder  string `json:"remote_folder"` // Subfolder name in the shared Drive/folder, e.g. "Valheim_OurWorld"
	AutoSync      bool   `json:"auto_sync"`
	AutoPull      bool   `json:"auto_pull"`
	KeepBackups   int    `json:"keep_backups"` // Max local safety backups to keep
	LastSyncedVer int    `json:"last_synced_ver"`
	LastLocalHash string `json:"last_local_hash"`
	LastSyncAt    string `json:"last_sync_at,omitempty"`
}

// GDriveConfig holds Google OAuth2 credentials and tokens.
type GDriveConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	FolderID     string `json:"folder_id"` // Shared Google Drive folder ID
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenExpiry  int64  `json:"token_expiry,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
}

// AppConfig is the root persistent configuration.
type AppConfig struct {
	mu             sync.RWMutex
	PlayerName     string        `json:"player_name"`
	Language       string        `json:"language"` // "ru" or "en"
	Backend        BackendType   `json:"backend"`
	LocalSyncDir   string        `json:"local_sync_dir"`
	GDrive         GDriveConfig  `json:"gdrive"`
	MinimizeToTray bool          `json:"minimize_to_tray"`
	StartMinimized bool          `json:"start_minimized"`
	PollIntervalMs int           `json:"poll_interval_ms"`
	Games          []GameProfile `json:"games"`
}

// AppDataDir returns %APPDATA%\SyncSave (or fallback to executable dir).
func AppDataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		exe, _ := os.Executable()
		dir = filepath.Dir(exe)
	}
	return filepath.Join(dir, "SyncSave")
}

// BackupsDir returns the local safety-backup directory.
func BackupsDir() string {
	return filepath.Join(AppDataDir(), "backups")
}

// ConfigPath returns the path to config.json.
func ConfigPath() string {
	return filepath.Join(AppDataDir(), "config.json")
}

// DefaultConfig returns sensible defaults for low-resource operation.
func DefaultConfig() *AppConfig {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "Player"
	}
	return &AppConfig{
		PlayerName:     hostname,
		Language:       "ru",
		Backend:        BackendGDrive,
		LocalSyncDir:   filepath.Join(AppDataDir(), "SharedSaves"),
		GDrive: GDriveConfig{
			FolderID: "1LfiBTS0x4vHFJJi9dK1H64Bj5zi5md8X",
		},
		MinimizeToTray: true,
		StartMinimized: false,
		PollIntervalMs: 5000,
		Games:          []GameProfile{},
	}
}

// Load reads config.json from disk, creating defaults if missing.
func Load() (*AppConfig, error) {
	p := ConfigPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			_ = cfg.Save()
			return cfg, nil
		}
		return nil, err
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.PollIntervalMs < 1500 {
		cfg.PollIntervalMs = 5000
	}
	if cfg.Language == "" {
		cfg.Language = "ru"
	}
	return cfg, nil
}

// Save writes config.json atomically.
func (c *AppConfig) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saveLocked()
}

func (c *AppConfig) saveLocked() error {
	dir := filepath.Dir(ConfigPath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := ConfigPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, ConfigPath())
}

// Snapshot returns a thread-safe copy of the config for reading.
func (c *AppConfig) Snapshot() AppConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cp := *c
	cp.Games = make([]GameProfile, len(c.Games))
	copy(cp.Games, c.Games)
	return cp
}

// Update mutates the config safely and persists it.
func (c *AppConfig) Update(fn func(*AppConfig)) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn(c)
	return c.saveLocked()
}
