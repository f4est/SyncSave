package archive

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DirectoryHash computes a deterministic SHA-256 hash of all files (relative paths + contents)
// in dir without loading entire files into memory.
func DirectoryHash(dir string) (string, int64, int, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", 0, 0, err
	}
	if !info.IsDir() {
		return "", 0, 0, fmt.Errorf("path is not a directory: %s", dir)
	}

	type fileEntry struct {
		rel  string
		full string
		size int64
	}
	var entries []fileEntry

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		// Skip temporary/lock files created by games while running
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".lock") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		entries = append(entries, fileEntry{
			rel:  filepath.ToSlash(rel),
			full: path,
			size: fi.Size(),
		})
		return nil
	})
	if err != nil {
		return "", 0, 0, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].rel < entries[j].rel
	})

	h := sha256.New()
	var totalSize int64
	buf := make([]byte, 32*1024)

	for _, e := range entries {
		totalSize += e.size
		_, _ = fmt.Fprintf(h, "%s:%d\n", e.rel, e.size)
		f, err := os.Open(e.full)
		if err != nil {
			return "", 0, 0, err
		}
		_, copyErr := io.CopyBuffer(h, f, buf)
		_ = f.Close()
		if copyErr != nil {
			return "", 0, 0, copyErr
		}
	}

	return hex.EncodeToString(h.Sum(nil)), totalSize, len(entries), nil
}

// PackDirectory compresses srcDir into dstZip using streaming Deflate and returns
// the resulting archive size and SHA-256 of the archive file.
func PackDirectory(srcDir, dstZip string) (int64, string, error) {
	if err := os.MkdirAll(filepath.Dir(dstZip), 0o755); err != nil {
		return 0, "", err
	}

	tmpZip := dstZip + ".tmp"
	f, err := os.Create(tmpZip)
	if err != nil {
		return 0, "", err
	}

	zw := zip.NewWriter(f)
	buf := make([]byte, 64*1024)

	walkErr := filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".lock") {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		hdr.Method = zip.Deflate

		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.CopyBuffer(w, srcFile, buf)
		_ = srcFile.Close()
		return copyErr
	})

	closeZipErr := zw.Close()
	closeFileErr := f.Close()

	if walkErr != nil {
		_ = os.Remove(tmpZip)
		return 0, "", walkErr
	}
	if closeZipErr != nil {
		_ = os.Remove(tmpZip)
		return 0, "", closeZipErr
	}
	if closeFileErr != nil {
		_ = os.Remove(tmpZip)
		return 0, "", closeFileErr
	}

	if err := os.Rename(tmpZip, dstZip); err != nil {
		_ = os.Remove(tmpZip)
		return 0, "", err
	}

	fi, err := os.Stat(dstZip)
	if err != nil {
		return 0, "", err
	}
	hash, err := FileSHA256(dstZip)
	if err != nil {
		return 0, "", err
	}
	return fi.Size(), hash, nil
}

// FileSHA256 computes the SHA-256 hex digest of a single file.
func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, 64*1024)
	if _, err := io.CopyBuffer(h, f, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// UnpackDirectory safely extracts srcZip into a staging directory, then replaces dstDir
// to prevent partial/corrupt save states.
func UnpackDirectory(srcZip, dstDir string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()

	stageDir := dstDir + ".syncsave_stage"
	_ = os.RemoveAll(stageDir)
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return err
	}

	buf := make([]byte, 64*1024)
	cleanStage := filepath.Clean(stageDir) + string(os.PathSeparator)

	for _, f := range r.File {
		target := filepath.Join(stageDir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), cleanStage) && filepath.Clean(target) != filepath.Clean(stageDir) {
			_ = os.RemoveAll(stageDir)
			return fmt.Errorf("unsafe zip entry path: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				_ = os.RemoveAll(stageDir)
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			_ = os.RemoveAll(stageDir)
			return err
		}
		rc, err := f.Open()
		if err != nil {
			_ = os.RemoveAll(stageDir)
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			_ = rc.Close()
			_ = os.RemoveAll(stageDir)
			return err
		}
		_, copyErr := io.CopyBuffer(out, rc, buf)
		_ = out.Close()
		_ = rc.Close()
		if copyErr != nil {
			_ = os.RemoveAll(stageDir)
			return copyErr
		}
	}

	// Swap old dir out and staged dir in
	oldDir := dstDir + ".syncsave_old"
	_ = os.RemoveAll(oldDir)
	if _, err := os.Stat(dstDir); err == nil {
		if err := os.Rename(dstDir, oldDir); err != nil {
			_ = os.RemoveAll(stageDir)
			return fmt.Errorf("cannot replace save folder (is the game still running?): %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dstDir), 0o755); err != nil {
		_ = os.Rename(oldDir, dstDir)
		_ = os.RemoveAll(stageDir)
		return err
	}
	if err := os.Rename(stageDir, dstDir); err != nil {
		_ = os.Rename(oldDir, dstDir)
		_ = os.RemoveAll(stageDir)
		return err
	}
	_ = os.RemoveAll(oldDir)
	return nil
}
