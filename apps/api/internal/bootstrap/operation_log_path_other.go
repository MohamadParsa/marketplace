//go:build !windows

package bootstrap

import "path/filepath"

func existingLogPath(path string) (string, error) { return filepath.EvalSymlinks(path) }
