package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands leading ~ or ~/ to the user's home directory.
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// DumpPayload atomically writes data to the target path with 0600 file permissions.
// If path is empty, it returns nil immediately without performing any action.
func DumpPayload(path string, data []byte) error {
	resolvedPath := ExpandPath(path)
	if resolvedPath == "" {
		return nil
	}

	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	base := filepath.Base(resolvedPath)
	tempFile, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	var success bool
	defer func() {
		_ = tempFile.Close()
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		return err
	}
	if err := tempFile.Sync(); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, resolvedPath); err != nil {
		return err
	}

	success = true
	return nil
}
