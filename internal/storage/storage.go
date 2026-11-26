package storage

import "context"

// SaveVersion represents one immutable snapshot of a co-op world save.
type SaveVersion struct {
	Version    int    `json:"version"`
	Author     string `json:"author"`
	Device     string `json:"device"`
	Timestamp  string `json:"timestamp"`
	Note       string `json:"note,omitempty"`
	DirHash    string `json:"dir_hash"`
	ArchiveSHA string `json:"archive_sha"`
	RawSize    int64  `json:"raw_size"`
	ZipSize    int64  `json:"zip_size"`
	FileCount  int    `json:"file_count"`
	FileName   string `json:"file_name"`
	RemoteID   string `json:"remote_id,omitempty"`
}

// HostLock indicates which player is currently hosting the shared world.
type HostLock struct {
	Holder    string `json:"holder"`
	Device    string `json:"device"`
	AcquiredAt string `json:"acquired_at"`
}

// WorldManifest is stored in each game's remote folder as manifest.json.
type WorldManifest struct {
	GameName       string        `json:"game_name"`
	LatestVersion  int           `json:"latest_version"`
	UpdatedAt      string        `json:"updated_at"`
	UpdatedBy      string        `json:"updated_by"`
	ActiveHostLock *HostLock     `json:"active_host_lock,omitempty"`
	Versions       []SaveVersion `json:"versions"`
}

// Backend abstracts online/shared save storage (Google Drive or Local Folder).
type Backend interface {
	Name() string
	CheckConnection(ctx context.Context) error
	GetManifest(ctx context.Context, remoteFolder string) (*WorldManifest, error)
	PutManifest(ctx context.Context, remoteFolder string, m *WorldManifest) error
	UploadArchive(ctx context.Context, remoteFolder, fileName, localZipPath string) (remoteID string, err error)
	DownloadArchive(ctx context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error
	ListRemoteWorlds(ctx context.Context) ([]string, error)
}
