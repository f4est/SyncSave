package lan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"syncsave/internal/config"
	"syncsave/internal/storage"
)

// Manager runs an optional LAN/VPN HTTP sync server so friends on the same network,
// Radmin VPN, ZeroTier, or Tailscale can push/pull saves directly.
type Manager struct {
	cfg *config.AppConfig
	mu  sync.Mutex
	srv *http.Server
}

func NewManager(cfg *config.AppConfig) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) SyncState() {
	snap := m.cfg.Snapshot()
	m.mu.Lock()
	defer m.mu.Unlock()

	if !snap.LAN.HostEnabled {
		if m.srv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = m.srv.Shutdown(ctx)
			cancel()
			m.srv = nil
		}
		return
	}

	if m.srv != nil {
		return
	}

	port := snap.LAN.Port
	if port <= 0 {
		port = 47821
	}

	mux := http.NewServeMux()
	auth := func(w http.ResponseWriter, r *http.Request) bool {
		pin := strings.TrimSpace(m.cfg.Snapshot().LAN.RoomPIN)
		if pin != "" && r.Header.Get("X-SyncSave-PIN") != pin {
			http.Error(w, "invalid PIN", http.StatusUnauthorized)
			return false
		}
		return true
	}
	local := func() *storage.LocalBackend {
		return storage.NewLocalBackend(m.cfg.Snapshot().LocalSyncDir)
	}

	mux.HandleFunc("/lan/ping", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	mux.HandleFunc("/lan/worlds", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		worlds, err := local().ListRemoteWorlds(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(worlds)
	})

	mux.HandleFunc("/lan/manifest", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		world := strings.TrimSpace(r.URL.Query().Get("world"))
		if world == "" || strings.ContainsAny(world, `/\`) || strings.Contains(world, "..") {
			http.Error(w, "invalid world", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			man, err := local().GetManifest(r.Context(), world)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if man == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(man)
		case http.MethodPut:
			var man storage.WorldManifest
			if err := json.NewDecoder(r.Body).Decode(&man); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := local().PutManifest(r.Context(), world, &man); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/lan/archive", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		world := strings.TrimSpace(r.URL.Query().Get("world"))
		file := filepath.Base(strings.TrimSpace(r.URL.Query().Get("file")))
		if world == "" || file == "" || file == "." || strings.ContainsAny(world, `/\`) || strings.Contains(world, "..") {
			http.Error(w, "invalid params", http.StatusBadRequest)
			return
		}
		dir := filepath.Join(m.cfg.Snapshot().LocalSyncDir, world)
		full := filepath.Join(dir, file)

		switch r.Method {
		case http.MethodGet:
			http.ServeFile(w, r, full)
		case http.MethodPut:
			if err := os.MkdirAll(dir, 0o755); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmp := full + ".tmp"
			out, err := os.Create(tmp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, copyErr := io.Copy(out, r.Body)
			_ = out.Close()
			if copyErr != nil {
				_ = os.Remove(tmp)
				http.Error(w, copyErr.Error(), http.StatusInternalServerError)
				return
			}
			if err := os.Rename(tmp, full); err != nil {
				_ = os.Remove(tmp)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: mux,
	}
	m.srv = srv
	go func() {
		_ = srv.ListenAndServe()
	}()
}

// LocalIPs returns active non-loopback IPv4 addresses (LAN, Radmin VPN, ZeroTier, etc.).
func LocalIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var ips []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}
