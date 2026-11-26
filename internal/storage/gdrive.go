package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"syncsave/internal/config"
)

const (
	gdriveFilesAPI  = "https://www.googleapis.com/drive/v3/files"
	gdriveUploadAPI = "https://www.googleapis.com/upload/drive/v3/files"
	googleTokenURL  = "https://oauth2.googleapis.com/token"
	googleAuthURL   = "https://accounts.google.com/o/oauth2/v2/auth"
	googleUserInfo  = "https://www.googleapis.com/oauth2/v2/userinfo"
	driveScope      = "https://www.googleapis.com/auth/drive email"
)

// GDriveBackend implements Backend using pure-Go Google Drive REST API v3.
// No heavy google.golang.org/api dependency is used, keeping the binary & RAM minimal.
type GDriveBackend struct {
	cfg        *config.AppConfig
	httpClient *http.Client
	mu         sync.Mutex
	folderCache map[string]string // remoteFolder name -> Drive folder ID
}

func NewGDriveBackend(cfg *config.AppConfig) *GDriveBackend {
	return &GDriveBackend{
		cfg:         cfg,
		httpClient:  &http.Client{Timeout: 120 * time.Second},
		folderCache: make(map[string]string),
	}
}

func (g *GDriveBackend) Name() string {
	return "gdrive"
}

// StartOAuthFlow launches a temporary loopback HTTP server, opens the user's browser
// to Google's consent screen, and stores the resulting access + refresh tokens.
func (g *GDriveBackend) StartOAuthFlow(ctx context.Context, clientID, clientSecret string) (string, error) {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return "", fmt.Errorf("Client ID и Client Secret обязательны")
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/oauth2callback", port)

	stateBytes := make([]byte, 16)
	_, _ = rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			errCh <- fmt.Errorf("invalid OAuth state")
			return
		}
		if e := r.URL.Query().Get("error"); e != "" {
			http.Error(w, "OAuth error: "+e, http.StatusBadRequest)
			errCh <- fmt.Errorf("google oauth error: %s", e)
			return
		}
		code := r.URL.Query().Get("code")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body style="background:#0f172a;color:#e2e8f0;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0"><div style="text-align:center"><h2>SyncSave: Google Drive подключен!</h2><p>Можете закрыть эту вкладку и вернуться в приложение.</p></div></body></html>`))
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent&state=%s",
		googleAuthURL,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(driveScope),
		url.QueryEscape(state),
	)

	// Open default browser on Windows
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", authURL).Start()

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(3 * time.Minute):
		return "", fmt.Errorf("время ожидания авторизации Google истекло (3 мин)")
	}

	form := url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(body))
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}

	email := g.fetchUserEmail(ctx, tok.AccessToken)
	expiry := time.Now().Unix() + tok.ExpiresIn - 60

	err = g.cfg.Update(func(c *config.AppConfig) {
		c.GDrive.ClientID = clientID
		c.GDrive.ClientSecret = clientSecret
		c.GDrive.AccessToken = tok.AccessToken
		if tok.RefreshToken != "" {
			c.GDrive.RefreshToken = tok.RefreshToken
		}
		c.GDrive.TokenExpiry = expiry
		if email != "" {
			c.GDrive.UserEmail = email
		}
	})
	return email, err
}

func (g *GDriveBackend) fetchUserEmail(ctx context.Context, token string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfo, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var info struct {
		Email string `json:"email"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&info)
	return info.Email
}

func (g *GDriveBackend) ensureToken(ctx context.Context) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	snap := g.cfg.Snapshot().GDrive
	if snap.AccessToken == "" && snap.RefreshToken == "" {
		return "", fmt.Errorf("Google Drive не подключен (войдите в настройках)")
	}
	if snap.AccessToken != "" && time.Now().Unix() < snap.TokenExpiry {
		return snap.AccessToken, nil
	}
	if snap.RefreshToken == "" {
		return "", fmt.Errorf("токен истёк и отсутствует refresh_token — переподключите Google Drive")
	}

	form := url.Values{
		"client_id":     {snap.ClientID},
		"client_secret": {snap.ClientSecret},
		"refresh_token": {snap.RefreshToken},
		"grant_type":    {"refresh_token"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	expiry := time.Now().Unix() + tok.ExpiresIn - 60
	_ = g.cfg.Update(func(c *config.AppConfig) {
		c.GDrive.AccessToken = tok.AccessToken
		c.GDrive.TokenExpiry = expiry
	})
	return tok.AccessToken, nil
}

func (g *GDriveBackend) rootFolderID() (string, error) {
	id := strings.TrimSpace(g.cfg.Snapshot().GDrive.FolderID)
	if id == "" {
		return "", fmt.Errorf("не указан ID общей папки Google Drive")
	}
	return id, nil
}

func (g *GDriveBackend) CheckConnection(ctx context.Context) error {
	tok, err := g.ensureToken(ctx)
	if err != nil {
		return err
	}
	folderID, err := g.rootFolderID()
	if err != nil {
		return err
	}
	u := fmt.Sprintf("%s/%s?fields=id,name,mimeType&supportsAllDrives=true", gdriveFilesAPI, url.PathEscape(folderID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("папка Google Drive недоступна (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

type driveFile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
}

func (g *GDriveBackend) findChild(ctx context.Context, parentID, name, mimeType string) (*driveFile, error) {
	tok, err := g.ensureToken(ctx)
	if err != nil {
		return nil, err
	}
	escapedName := strings.ReplaceAll(name, `'`, `\'`)
	q := fmt.Sprintf("'%s' in parents and name = '%s' and trashed = false", parentID, escapedName)
	if mimeType != "" {
		q += fmt.Sprintf(" and mimeType = '%s'", mimeType)
	}
	u := fmt.Sprintf("%s?q=%s&fields=files(id,name,mimeType)&supportsAllDrives=true&includeItemsFromAllDrives=true",
		gdriveFilesAPI, url.QueryEscape(q))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("drive list error (%d): %s", resp.StatusCode, string(b))
	}
	var out struct {
		Files []driveFile `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Files) == 0 {
		return nil, nil
	}
	return &out.Files[0], nil
}

func (g *GDriveBackend) ensureWorldFolder(ctx context.Context, remoteFolder string, createIfMissing bool) (string, error) {
	g.mu.Lock()
	cached, ok := g.folderCache[remoteFolder]
	g.mu.Unlock()
	if ok && cached != "" {
		return cached, nil
	}

	rootID, err := g.rootFolderID()
	if err != nil {
		return "", err
	}
	f, err := g.findChild(ctx, rootID, remoteFolder, "application/vnd.google-apps.folder")
	if err != nil {
		return "", err
	}
	if f != nil {
		g.mu.Lock()
		g.folderCache[remoteFolder] = f.ID
		g.mu.Unlock()
		return f.ID, nil
	}
	if !createIfMissing {
		return "", nil
	}

	tok, err := g.ensureToken(ctx)
	if err != nil {
		return "", err
	}
	meta := map[string]any{
		"name":     remoteFolder,
		"mimeType": "application/vnd.google-apps.folder",
		"parents":  []string{rootID},
	}
	metaBytes, _ := json.Marshal(meta)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gdriveFilesAPI+"?supportsAllDrives=true", bytes.NewReader(metaBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create folder failed (%d): %s", resp.StatusCode, string(b))
	}
	var created driveFile
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", err
	}
	g.mu.Lock()
	g.folderCache[remoteFolder] = created.ID
	g.mu.Unlock()
	return created.ID, nil
}

func (g *GDriveBackend) GetManifest(ctx context.Context, remoteFolder string) (*WorldManifest, error) {
	folderID, err := g.ensureWorldFolder(ctx, remoteFolder, false)
	if err != nil {
		return nil, err
	}
	if folderID == "" {
		return nil, nil
	}
	mf, err := g.findChild(ctx, folderID, "manifest.json", "")
	if err != nil || mf == nil {
		return nil, err
	}

	tok, err := g.ensureToken(ctx)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/%s?alt=media&supportsAllDrives=true", gdriveFilesAPI, url.PathEscape(mf.ID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download manifest failed (%d): %s", resp.StatusCode, string(b))
	}
	var m WorldManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (g *GDriveBackend) PutManifest(ctx context.Context, remoteFolder string, m *WorldManifest) error {
	folderID, err := g.ensureWorldFolder(ctx, remoteFolder, true)
	if err != nil {
		return err
	}
	existing, err := g.findChild(ctx, folderID, "manifest.json", "")
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if existing != nil {
		_, err = g.multipartUpload(ctx, existing.ID, "", "manifest.json", "application/json", bytes.NewReader(data))
		return err
	}
	_, err = g.multipartUpload(ctx, "", folderID, "manifest.json", "application/json", bytes.NewReader(data))
	return err
}

func (g *GDriveBackend) UploadArchive(ctx context.Context, remoteFolder, fileName, localZipPath string) (string, error) {
	folderID, err := g.ensureWorldFolder(ctx, remoteFolder, true)
	if err != nil {
		return "", err
	}
	f, err := os.Open(localZipPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return g.multipartUpload(ctx, "", folderID, fileName, "application/zip", f)
}

func (g *GDriveBackend) multipartUpload(ctx context.Context, existingFileID, parentID, name, contentType string, r io.Reader) (string, error) {
	tok, err := g.ensureToken(ctx)
	if err != nil {
		return "", err
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		var meta map[string]any
		if existingFileID != "" {
			meta = map[string]any{"name": name}
		} else {
			meta = map[string]any{"name": name, "parents": []string{parentID}}
		}
		metaBytes, _ := json.Marshal(meta)

		mh := textproto.MIMEHeader{}
		mh.Set("Content-Type", "application/json; charset=UTF-8")
		p1, err := mw.CreatePart(mh)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := p1.Write(metaBytes); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		fh := textproto.MIMEHeader{}
		fh.Set("Content-Type", contentType)
		p2, err := mw.CreatePart(fh)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		buf := make([]byte, 64*1024)
		if _, err := io.CopyBuffer(p2, r, buf); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = pw.CloseWithError(mw.Close())
	}()

	method := http.MethodPost
	u := gdriveUploadAPI + "?uploadType=multipart&supportsAllDrives=true"
	if existingFileID != "" {
		method = http.MethodPatch
		u = fmt.Sprintf("%s/%s?uploadType=multipart&supportsAllDrives=true", gdriveUploadAPI, url.PathEscape(existingFileID))
	}

	req, err := http.NewRequestWithContext(ctx, method, u, pr)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "multipart/related; boundary="+mw.Boundary())

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("drive upload failed (%d): %s", resp.StatusCode, string(body))
	}
	var out driveFile
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (g *GDriveBackend) DownloadArchive(ctx context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error {
	fileID := ver.RemoteID
	if fileID == "" {
		folderID, err := g.ensureWorldFolder(ctx, remoteFolder, false)
		if err != nil {
			return err
		}
		if folderID == "" {
			return fmt.Errorf("папка мира %s не найдена на Google Drive", remoteFolder)
		}
		f, err := g.findChild(ctx, folderID, ver.FileName, "")
		if err != nil {
			return err
		}
		if f == nil {
			return fmt.Errorf("архив %s не найден на Google Drive", ver.FileName)
		}
		fileID = f.ID
	}

	tok, err := g.ensureToken(ctx)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("%s/%s?alt=media&supportsAllDrives=true", gdriveFilesAPI, url.PathEscape(fileID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("drive download failed (%d): %s", resp.StatusCode, string(b))
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

func (g *GDriveBackend) ListRemoteWorlds(ctx context.Context) ([]string, error) {
	tok, err := g.ensureToken(ctx)
	if err != nil {
		return nil, err
	}
	rootID, err := g.rootFolderID()
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false", rootID)
	u := fmt.Sprintf("%s?q=%s&fields=files(id,name)&supportsAllDrives=true&includeItemsFromAllDrives=true",
		gdriveFilesAPI, url.QueryEscape(q))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list worlds failed (%d): %s", resp.StatusCode, string(b))
	}
	var out struct {
		Files []driveFile `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(out.Files))
	for _, f := range out.Files {
		names = append(names, f.Name)
		g.mu.Lock()
		g.folderCache[f.Name] = f.ID
		g.mu.Unlock()
	}
	return names, nil
}
