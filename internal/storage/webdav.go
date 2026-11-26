package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// WebDAVBackend implements Backend over WebDAV (Yandex.Disk, Nextcloud, OwnCloud, NAS).
type WebDAVBackend struct {
	BaseURL    string
	Username   string
	Password   string
	BasePath   string
	httpClient *http.Client
}

func NewWebDAVBackend(baseURL, username, password, basePath string) *WebDAVBackend {
	return &WebDAVBackend{
		BaseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		Username:   strings.TrimSpace(username),
		Password:   password,
		BasePath:   strings.Trim(strings.TrimSpace(basePath), "/"),
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (w *WebDAVBackend) Name() string {
	return "webdav"
}

func (w *WebDAVBackend) buildURL(parts ...string) (string, error) {
	if w.BaseURL == "" {
		return "", fmt.Errorf("не указан URL WebDAV-сервера")
	}
	u, err := url.Parse(w.BaseURL)
	if err != nil {
		return "", fmt.Errorf("некорректный URL WebDAV: %w", err)
	}
	all := []string{strings.TrimRight(u.Path, "/")}
	if w.BasePath != "" {
		all = append(all, w.BasePath)
	}
	for _, p := range parts {
		clean := strings.Trim(p, "/")
		if clean != "" {
			all = append(all, clean)
		}
	}
	u.Path = path.Join(all...)
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	return u.String(), nil
}

func (w *WebDAVBackend) doReq(ctx context.Context, method, rawURL string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if w.Username != "" {
		req.SetBasicAuth(w.Username, w.Password)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return w.httpClient.Do(req)
}

func (w *WebDAVBackend) ensureDir(ctx context.Context, parts ...string) error {
	acc := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		acc = append(acc, p)
		u, err := w.buildURL(acc...)
		if err != nil {
			return err
		}
		resp, err := w.doReq(ctx, "MKCOL", u, nil, "")
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		// 201 Created, 405 Method Not Allowed (already exists), 301/409 ok if parent exists
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("WebDAV MKCOL %s вернул %d", u, resp.StatusCode)
		}
	}
	return nil
}

func (w *WebDAVBackend) CheckConnection(ctx context.Context) error {
	if err := w.ensureDir(ctx); err != nil {
		return err
	}
	u, err := w.buildURL()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", u, nil)
	if err != nil {
		return err
	}
	if w.Username != "" {
		req.SetBasicAuth(w.Username, w.Password)
	}
	req.Header.Set("Depth", "0")
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("ошибка авторизации WebDAV (%d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 && resp.StatusCode != 405 {
		return fmt.Errorf("WebDAV недоступен (%d)", resp.StatusCode)
	}
	return nil
}

func (w *WebDAVBackend) GetManifest(ctx context.Context, remoteFolder string) (*WorldManifest, error) {
	u, err := w.buildURL(remoteFolder, "manifest.json")
	if err != nil {
		return nil, err
	}
	resp, err := w.doReq(ctx, http.MethodGet, u, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("WebDAV GET manifest (%d)", resp.StatusCode)
	}
	var m WorldManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (w *WebDAVBackend) PutManifest(ctx context.Context, remoteFolder string, m *WorldManifest) error {
	if err := w.ensureDir(ctx, remoteFolder); err != nil {
		return err
	}
	u, err := w.buildURL(remoteFolder, "manifest.json")
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	resp, err := w.doReq(ctx, http.MethodPut, u, bytes.NewReader(data), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("WebDAV PUT manifest (%d)", resp.StatusCode)
	}
	return nil
}

func (w *WebDAVBackend) UploadArchive(ctx context.Context, remoteFolder, fileName, localZipPath string) (string, error) {
	if err := w.ensureDir(ctx, remoteFolder); err != nil {
		return "", err
	}
	u, err := w.buildURL(remoteFolder, fileName)
	if err != nil {
		return "", err
	}
	f, err := os.Open(localZipPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	resp, err := w.doReq(ctx, http.MethodPut, u, f, "application/zip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("WebDAV PUT archive (%d)", resp.StatusCode)
	}
	return u, nil
}

func (w *WebDAVBackend) DownloadArchive(ctx context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error {
	u, err := w.buildURL(remoteFolder, ver.FileName)
	if err != nil {
		return err
	}
	resp, err := w.doReq(ctx, http.MethodGet, u, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("WebDAV GET archive (%d)", resp.StatusCode)
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

func (w *WebDAVBackend) ListRemoteWorlds(ctx context.Context) ([]string, error) {
	u, err := w.buildURL()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", u, nil)
	if err != nil {
		return nil, err
	}
	if w.Username != "" {
		req.SetBasicAuth(w.Username, w.Password)
	}
	req.Header.Set("Depth", "1")
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("WebDAV PROPFIND (%d)", resp.StatusCode)
	}
	var ms struct {
		Responses []struct {
			Href string `xml:"href"`
		} `xml:"response"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&ms); err != nil {
		return nil, err
	}
	baseParsed, _ := url.Parse(u)
	basePath := strings.TrimRight(baseParsed.Path, "/")
	var worlds []string
	for _, r := range ms.Responses {
		hp := strings.TrimRight(r.Href, "/")
		if p, err := url.PathUnescape(hp); err == nil {
			hp = p
		}
		if hp == basePath || !strings.HasPrefix(hp, basePath+"/") {
			continue
		}
		name := strings.TrimPrefix(hp, basePath+"/")
		if name != "" && !strings.Contains(name, "/") {
			worlds = append(worlds, name)
		}
	}
	sort.Strings(worlds)
	return worlds, nil
}
