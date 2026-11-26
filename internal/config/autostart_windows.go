//go:build windows

package config

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValueName = "SyncSave"

// SetAutostart enables or disables launching SyncSave when Windows starts.
func SetAutostart(enabled, minimized bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if !enabled {
		err := k.DeleteValue(runValueName)
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf(`"%s"`, exe)
	if minimized {
		cmd += " --minimized"
	}
	return k.SetStringValue(runValueName, cmd)
}

// IsAutostartEnabled checks whether the Windows Run registry entry exists.
func IsAutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(runValueName)
	return err == nil
}
