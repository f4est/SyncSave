package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// LocalBackend stores shared worlds in a local or OS-synced directory
// (e.g. Google Drive Desktop, OneDrive, Yandex Disk, or LAN share).
type LocalBackend struct {
	RootDir string
}

func NewLocalBackend(rootDir string) *LocalBackend {
	return &LocalBackend{RootDir: rootDir}
}

func (b *LocalBackend) Name() string {
	return "local"
}

func (b *LocalBackend) CheckConnection(_ context.Context) error {
	if b.RootDir == "" {
		return fmt.Errorf("local sync directory is not configured")
	}
	return os.MkdirAll(b.RootDir, 0o755)
}

func (b *LocalBackend) worldDir(remoteFolder string) string {
	return filepath.Join(b.RootDir, filepath.Clean(remoteFolder))
}

func (b *LocalBackend) GetManifest(_ context.Context, remoteFolder string) (*WorldManifest, error) {
	p := filepath.Join(b.worldDir(remoteFolder), "manifest.json")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var m WorldManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (b *LocalBackend) PutManifest(_ context.Context, remoteFolder string, m *WorldManifest) error {
	dir := b.worldDir(remoteFolder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(dir, "manifest.json")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (b *LocalBackend) UploadArchive(_ context.Context, remoteFolder, fileName, localZipPath string) (string, error) {
	dir := b.worldDir(remoteFolder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, fileName)
	tmp := dst + ".tmp"

	src, err := os.Open(localZipPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	buf := make([]byte, 64*1024)
	_, copyErr := io.CopyBuffer(out, src, buf)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return "", closeErr
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return dst, nil
}

func (b *LocalBackend) DownloadArchive(_ context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error {
	srcPath := filepath.Join(b.worldDir(remoteFolder), ver.FileName)
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dstZipPath), 0o755); err != nil {
		return err
	}
	tmp := dstZipPath + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	buf := make([]byte, 64*1024)
	_, copyErr := io.CopyBuffer(out, src, buf)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	return os.Rename(tmp, dstZipPath)
}

func (b *LocalBackend) ListRemoteWorlds(_ context.Context) ([]string, error) {
	if err := os.MkdirAll(b.RootDir, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(b.RootDir)
	if err != nil {
		return nil, err
	}
	var worlds []string
	for _, e := range entries {
		if e.IsDir() {
			worlds = append(worlds, e.Name())
		}
	}
	sort.Strings(worlds)
	return worlds, nil
}
