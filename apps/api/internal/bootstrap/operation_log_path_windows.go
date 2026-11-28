//go:build windows

package bootstrap

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// A directory handle follows junctions as well as symlinks. EvalSymlinks alone
// can preserve a junction's spelling on Windows and miss a public directory.
func existingLogPath(path string) (string, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	handle, err := windows.CreateFile(name, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(handle)
	buffer := make([]uint16, 1024)
	for {
		n, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
		if err != nil {
			return "", err
		}
		if n >= uint32(len(buffer)) {
			buffer = make([]uint16, n+1)
			continue
		}
		resolved := windows.UTF16ToString(buffer[:n])
		if strings.HasPrefix(resolved, `\\?\UNC\`) {
			resolved = `\\` + strings.TrimPrefix(resolved, `\\?\UNC\`)
		} else {
			resolved = strings.TrimPrefix(resolved, `\\?\`)
		}
		return filepath.Clean(resolved), nil
	}
}
