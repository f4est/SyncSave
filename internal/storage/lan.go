package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LANBackend connects to another SyncSave instance over LAN, Radmin VPN, ZeroTier, or Tailscale.
type LANBackend struct {
	PeerURL    string
	RoomPIN    string
	httpClient *http.Client
}

func NewLANBackend(peerURL, roomPIN string) *LANBackend {
	u := strings.TrimRight(strings.TrimSpace(peerURL), "/")
	if u != "" && !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}
	return &LANBackend{
		PeerURL:    u,
		RoomPIN:    strings.TrimSpace(roomPIN),
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (l *LANBackend) Name() string {
	return "lan"
}

func (l *LANBackend) doReq(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	if l.PeerURL == "" {
		return nil, fmt.Errorf("укажите IP/URL хоста (например http://26.10.20.30:47821)")
	}
	req, err := http.NewRequestWithContext(ctx, method, l.PeerURL+path, body)
	if err != nil {
		return nil, err
	}
	if l.RoomPIN != "" {
		req.Header.Set("X-SyncSave-PIN", l.RoomPIN)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return l.httpClient.Do(req)
}

func (l *LANBackend) CheckConnection(ctx context.Context) error {
	resp, err := l.doReq(ctx, http.MethodGet, "/lan/ping", nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("LAN-сервер вернул ошибку (%d)", resp.StatusCode)
	}
	return nil
}

func (l *LANBackend) GetManifest(ctx context.Context, remoteFolder string) (*WorldManifest, error) {
	resp, err := l.doReq(ctx, http.MethodGet, "/lan/manifest?world="+url.QueryEscape(remoteFolder), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("LAN manifest (%d)", resp.StatusCode)
	}
	var m WorldManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (l *LANBackend) PutManifest(ctx context.Context, remoteFolder string, m *WorldManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	resp, err := l.doReq(ctx, http.MethodPut, "/lan/manifest?world="+url.QueryEscape(remoteFolder), bytes.NewReader(data), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("LAN PUT manifest (%d)", resp.StatusCode)
	}
	return nil
}

func (l *LANBackend) UploadArchive(ctx context.Context, remoteFolder, fileName, localZipPath string) (string, error) {
	f, err := os.Open(localZipPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	p := fmt.Sprintf("/lan/archive?world=%s&file=%s", url.QueryEscape(remoteFolder), url.QueryEscape(fileName))
	resp, err := l.doReq(ctx, http.MethodPut, p, f, "application/zip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("LAN upload (%d)", resp.StatusCode)
	}
	return fileName, nil
}

func (l *LANBackend) DownloadArchive(ctx context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error {
	p := fmt.Sprintf("/lan/archive?world=%s&file=%s", url.QueryEscape(remoteFolder), url.QueryEscape(ver.FileName))
	resp, err := l.doReq(ctx, http.MethodGet, p, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("LAN download (%d)", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(dstZipPath), 0o755); err != nil {
		return err
	}
	tmp := dstZipPath + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	buf := make([]byte, 64*1024)
	_, copyErr := io.CopyBuffer(out, resp.Body, buf)
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

func (l *LANBackend) ListRemoteWorlds(ctx context.Context) ([]string, error) {
	resp, err := l.doReq(ctx, http.MethodGet, "/lan/worlds", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("LAN list worlds (%d)", resp.StatusCode)
	}
	var worlds []string
	if err := json.NewDecoder(resp.Body).Decode(&worlds); err != nil {
		return nil, err
	}
	return worlds, nil
}
