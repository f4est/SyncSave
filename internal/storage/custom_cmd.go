package storage

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CustomCmdBackend wraps a local staging directory with user-defined shell commands
// for Pull (before reading) and Push (after writing), supporting rclone, git, rsync, etc.
type CustomCmdBackend struct {
	local   *LocalBackend
	PullCmd string
	PushCmd string
}

func NewCustomCmdBackend(syncDir, pullCmd, pushCmd string) *CustomCmdBackend {
	return &CustomCmdBackend{
		local:   NewLocalBackend(syncDir),
		PullCmd: strings.TrimSpace(pullCmd),
		PushCmd: strings.TrimSpace(pushCmd),
	}
}

func (c *CustomCmdBackend) Name() string {
	return "custom"
}

func (c *CustomCmdBackend) runCmd(ctx context.Context, tmpl, remoteFolder string) error {
	if tmpl == "" {
		return nil
	}
	cmdStr := strings.NewReplacer(
		"{sync_dir}", c.local.RootDir,
		"{remote_folder}", remoteFolder,
	).Replace(tmpl)

	cmd := exec.CommandContext(ctx, "cmd", "/C", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("custom command failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (c *CustomCmdBackend) CheckConnection(ctx context.Context) error {
	if err := c.local.CheckConnection(ctx); err != nil {
		return err
	}
	return c.runCmd(ctx, c.PullCmd, "")
}

func (c *CustomCmdBackend) GetManifest(ctx context.Context, remoteFolder string) (*WorldManifest, error) {
	if err := c.runCmd(ctx, c.PullCmd, remoteFolder); err != nil {
		return nil, err
	}
	return c.local.GetManifest(ctx, remoteFolder)
}

func (c *CustomCmdBackend) PutManifest(ctx context.Context, remoteFolder string, m *WorldManifest) error {
	if err := c.local.PutManifest(ctx, remoteFolder, m); err != nil {
		return err
	}
	return c.runCmd(ctx, c.PushCmd, remoteFolder)
}

func (c *CustomCmdBackend) UploadArchive(ctx context.Context, remoteFolder, fileName, localZipPath string) (string, error) {
	return c.local.UploadArchive(ctx, remoteFolder, fileName, localZipPath)
}

func (c *CustomCmdBackend) DownloadArchive(ctx context.Context, remoteFolder string, ver SaveVersion, dstZipPath string) error {
	if err := c.runCmd(ctx, c.PullCmd, remoteFolder); err != nil {
		return err
	}
	return c.local.DownloadArchive(ctx, remoteFolder, ver, dstZipPath)
}

func (c *CustomCmdBackend) ListRemoteWorlds(ctx context.Context) ([]string, error) {
	_ = c.runCmd(ctx, c.PullCmd, "")
	return c.local.ListRemoteWorlds(ctx)
}
