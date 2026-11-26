package syncer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"syncsave/internal/config"
	"syncsave/internal/storage"
)

func TestCoopPushPullConflictAndRollback(t *testing.T) {
	tmp := t.TempDir()
	sharedDir := filepath.Join(tmp, "SharedCloud")
	playerASaves := filepath.Join(tmp, "PlayerA_Saves")
	playerBSaves := filepath.Join(tmp, "PlayerB_Saves")

	if err := os.MkdirAll(playerASaves, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(playerASaves, "world.db"), []byte("v1-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfgA := config.DefaultConfig()
	cfgA.PlayerName = "Alice"
	cfgA.Backend = config.BackendLocal
	cfgA.LocalSyncDir = sharedDir
	cfgA.Games = []config.GameProfile{{
		ID:           "g1",
		Name:         "Valheim",
		LocalPath:    playerASaves,
		RemoteFolder: "Valheim_Squad",
		KeepBackups:  3,
	}}

	cfgB := config.DefaultConfig()
	cfgB.PlayerName = "Bob"
	cfgB.Backend = config.BackendLocal
	cfgB.LocalSyncDir = sharedDir
	cfgB.Games = []config.GameProfile{{
		ID:           "g1",
		Name:         "Valheim",
		LocalPath:    playerBSaves,
		RemoteFolder: "Valheim_Squad",
		KeepBackups:  3,
	}}

	engA := NewEngine(cfgA, storage.NewGDriveBackend(cfgA), nil)
	engB := NewEngine(cfgB, storage.NewGDriveBackend(cfgB), nil)
	ctx := context.Background()

	// 1. Alice pushes v1
	v1, err := engA.Push(ctx, "g1", "Initial world", false)
	if err != nil {
		t.Fatalf("Alice push v1 failed: %v", err)
	}
	if v1.Version != 1 || v1.Author != "Alice" {
		t.Fatalf("unexpected v1: %+v", v1)
	}

	// 2. Bob pulls v1
	pulled, err := engB.Pull(ctx, "g1", 0, false)
	if err != nil {
		t.Fatalf("Bob pull v1 failed: %v", err)
	}
	if pulled.Version != 1 {
		t.Fatalf("expected pulled v1, got %d", pulled.Version)
	}
	got, _ := os.ReadFile(filepath.Join(playerBSaves, "world.db"))
	if string(got) != "v1-data" {
		t.Fatalf("expected v1-data, got %q", string(got))
	}

	// 3. Bob acquires host lock, updates save, and pushes v2
	if err := engB.SetHostLock(ctx, "g1", true); err != nil {
		t.Fatalf("Bob lock failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(playerBSaves, "world.db"), []byte("v2-boss-defeated"), 0o644); err != nil {
		t.Fatal(err)
	}
	v2, err := engB.Push(ctx, "g1", "Beat Elder", false)
	if err != nil {
		t.Fatalf("Bob push v2 failed: %v", err)
	}
	if v2.Version != 2 {
		t.Fatalf("expected v2, got %d", v2.Version)
	}
	if err := engB.SetHostLock(ctx, "g1", false); err != nil {
		t.Fatalf("Bob unlock failed: %v", err)
	}

	// 4. Alice modifies locally without pulling v2 -> Push must be blocked (conflict protection)
	if err := os.WriteFile(filepath.Join(playerASaves, "world.db"), []byte("alice-offline-edit"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := engA.Push(ctx, "g1", "Should fail", false); err == nil {
		t.Fatal("expected conflict error when Alice pushes without pulling v2")
	}

	stA := engA.GetGameStatus(ctx, "g1")
	if stA.State != StateConflict {
		t.Fatalf("expected StateConflict, got %s", stA.State)
	}

	// 5. Bob rolls back to v1
	if _, err := engB.Pull(ctx, "g1", 1, true); err != nil {
		t.Fatalf("Bob rollback to v1 failed: %v", err)
	}
	gotRollback, _ := os.ReadFile(filepath.Join(playerBSaves, "world.db"))
	if string(gotRollback) != "v1-data" {
		t.Fatalf("expected v1-data after rollback, got %q", string(gotRollback))
	}
}
