package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// S3Backend implements Backend for S3-compatible object storage
// (Cloudflare R2, MinIO, Backblaze B2, AWS S3, Selectel, Yandex Object Storage).
type S3Backend struct {
	Endpoint   string
	Region     string
	Bucket     string
	AccessKey  string
	SecretKey  string
	Prefix     string
	httpClient *http.Client
}

func NewS3Backend(endpoint, region, bucket, accessKey, secretKey, prefix string) *S3Backend {
	if region == "" {
		region = "auto"
	}
	return &S3Backend{
		Endpoint:   strings.TrimRight(strings.TrimSpace(endpoint), "/"),
		Region:     strings.TrimSpace(region),
		Bucket:     strings.TrimSpace(bucket),
		AccessKey:  strings.TrimSpace(accessKey),
		SecretKey:  strings.TrimSpace(secretKey),
		Prefix:     strings.Trim(strings.TrimSpace(prefix), "/"),
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (s *S3Backend) Name() string {
	return "s3"
}

func (s *S3Backend) objectKey(parts ...string) string {
	all := make([]string, 0, len(parts)+1)
	if s.Prefix != "" {
		all = append(all, s.Prefix)
	}
	for _, p := range parts {
		clean := strings.Trim(p, "/")
		if clean != "" {
			all = append(all, clean)
		}
	}
	return strings.Join(all, "/")
}

func (s *S3Backend) signedRequest(ctx context.Context, method, key, query string, body []byte, contentType string) (*http.Response, error) {
	if s.Endpoint == "" || s.Bucket == "" || s.AccessKey == "" || s.SecretKey == "" {
		return nil, fmt.Errorf("заполните Endpoint, Bucket, Access Key и Secret Key для S3")
	}

	u, err := url.Parse(s.Endpoint)
	if err != nil {
		return nil, err
	}
	pathPart := "/" + s.Bucket
	if key != "" {
		pathPart += "/" + key
	}
	u.Path = pathPart
	u.RawQuery = query

	payloadHash := sha256Hex(body)
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", u.Host)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n", u.Host, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalReq := strings.Join([]string{
		method,
		u.EscapedPath(),
		query,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	scope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, s.Region)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalReq)),
	}, "\n")

	kDate := hmacSHA256([]byte("AWS4"+s.SecretKey), dateStamp)
	kRegion := hmacSHA256(kDate, s.Region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))

	auth := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.AccessKey, scope, signedHeaders, signature)
	req.Header.Set("Authorization", auth)

	return s.httpClient.Do(req)
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

func (s *S3Backend) CheckConnection(ctx context.Context) error {
	resp, err := s.signedRequest(ctx, http.MethodGet, "", "list-type=2&max-keys=1", nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("S3 ошибка (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

func (s *S3Backend) GetManifest(ctx context.Context, remoteFolder string) (*WorldManifest, error) {
	key := s.objectKey(remoteFolder, "manifest.json")
	resp, err := s.signedRequest(ctx, http.MethodGet, key, "", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("S3 GET manifest (%d): %s", resp.StatusCode, string(b))
	}
	var m WorldManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *S3Backend) PutManifest(ctx context.Context, remoteFolder string, m *WorldManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	key := s.objectKey(remoteFolder, "manifest.json")
	resp, err := s.signedRequest(ctx, http.MethodPut, key, "", data, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("S3 PUT manifest (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}

func (s *S3Backend) UploadArchive(ctx context.Context, remoteFolder, fileName, localZipPath string) (string, error) {
	data, err := os.ReadFile(localZipPath)
	if err != nil {
		return "", err
	}
	key := s.objectKey(remoteFolder, fileName)
	resp, err := s.signedRequest(ctx, http.MethodPut, key, "", data, "application/zip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("S3 PUT archive (%d): %s", resp.StatusCode, string(b))
	}
	return key, nil
}

func (s *S3Backend) DownloadArchive(ctx context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error {
	key := s.objectKey(remoteFolder, ver.FileName)
	resp, err := s.signedRequest(ctx, http.MethodGet, key, "", nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("S3 GET archive (%d): %s", resp.StatusCode, string(b))
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

func (s *S3Backend) ListRemoteWorlds(ctx context.Context) ([]string, error) {
	q := "delimiter=%2F&list-type=2"
	if s.Prefix != "" {
		q += "&prefix=" + url.QueryEscape(s.Prefix+"/")
	}
	resp, err := s.signedRequest(ctx, http.MethodGet, "", q, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("S3 list worlds (%d)", resp.StatusCode)
	}
	var out struct {
		CommonPrefixes []struct {
			Prefix string `xml:"Prefix"`
		} `xml:"CommonPrefixes"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	worlds := make([]string, 0, len(out.CommonPrefixes))
	for _, cp := range out.CommonPrefixes {
		name := strings.Trim(cp.Prefix, "/")
		if s.Prefix != "" {
			name = strings.TrimPrefix(name, s.Prefix+"/")
		}
		if name != "" {
			worlds = append(worlds, name)
		}
	}
	sort.Strings(worlds)
	return worlds, nil
}
