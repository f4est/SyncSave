package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitBackend syncs saves via a private GitHub / GitLab / Gitea repository.
type GitBackend struct {
	RepoURL  string
	Branch   string
	LocalDir string
	local    *LocalBackend
}

func NewGitBackend(repoURL, branch, localDir string) *GitBackend {
	if strings.TrimSpace(branch) == "" {
		branch = "main"
	}
	return &GitBackend{
		RepoURL:  strings.TrimSpace(repoURL),
		Branch:   strings.TrimSpace(branch),
		LocalDir: filepath.Clean(localDir),
		local:    NewLocalBackend(localDir),
	}
}

func (g *GitBackend) Name() string {
	return "git"
}

func (g *GitBackend) runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func (g *GitBackend) ensureRepo(ctx context.Context) error {
	if g.RepoURL == "" {
		return fmt.Errorf("укажите URL Git-репозитория")
	}
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("Git не установлен на ПК (скачайте git-scm.com)")
	}
	gitDir := filepath.Join(g.LocalDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(g.LocalDir), 0o755); err != nil {
			return err
		}
		_ = os.RemoveAll(g.LocalDir)
		if _, err := g.runGit(ctx, "", "clone", "-b", g.Branch, g.RepoURL, g.LocalDir); err != nil {
			// Fallback if repo is empty
			if _, err2 := g.runGit(ctx, "", "clone", g.RepoURL, g.LocalDir); err2 != nil {
				return err
			}
		}
		return nil
	}
	_, _ = g.runGit(ctx, g.LocalDir, "pull", "--rebase", "origin", g.Branch)
	return nil
}

func (g *GitBackend) CheckConnection(ctx context.Context) error {
	return g.ensureRepo(ctx)
}

func (g *GitBackend) GetManifest(ctx context.Context, remoteFolder string) (*WorldManifest, error) {
	if err := g.ensureRepo(ctx); err != nil {
		return nil, err
	}
	return g.local.GetManifest(ctx, remoteFolder)
}

func (g *GitBackend) PutManifest(ctx context.Context, remoteFolder string, m *WorldManifest) error {
	if err := g.local.PutManifest(ctx, remoteFolder, m); err != nil {
		return err
	}
	if _, err := g.runGit(ctx, g.LocalDir, "add", "."); err != nil {
		return err
	}
	msg := fmt.Sprintf("SyncSave: %s v%d (%s)", remoteFolder, m.LatestVersion, m.UpdatedBy)
	_, _ = g.runGit(ctx, g.LocalDir, "-c", "user.name=SyncSave", "-c", "user.email=syncsave@local", "commit", "-m", msg)
	_, err := g.runGit(ctx, g.LocalDir, "push", "origin", "HEAD:"+g.Branch)
	return err
}

func (g *GitBackend) UploadArchive(ctx context.Context, remoteFolder, fileName, localZipPath string) (string, error) {
	return g.local.UploadArchive(ctx, remoteFolder, fileName, localZipPath)
}

func (g *GitBackend) DownloadArchive(ctx context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error {
	if err := g.ensureRepo(ctx); err != nil {
		return err
	}
	return g.local.DownloadArchive(ctx, remoteFolder, ver, dstZipPath)
}

func (g *GitBackend) ListRemoteWorlds(ctx context.Context) ([]string, error) {
	if err := g.ensureRepo(ctx); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(g.LocalDir)
	if err != nil {
		return nil, err
	}
	var worlds []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != ".git" {
			worlds = append(worlds, e.Name())
		}
	}
	return worlds, nil
}
