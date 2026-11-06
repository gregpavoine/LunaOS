//go:build linux

package core

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// OwnerSocketListener is a mode-restricted local API listener. Its trust boundary
// comes from the socket's owner/group permissions, not an HTTP bearer secret.
type OwnerSocketListener struct {
	*net.UnixListener
	path string
	info os.FileInfo
}

func ListenOwnerSocket(path string) (*OwnerSocketListener, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, fmt.Errorf("owner API socket path must be absolute and clean")
	}
	parent := filepath.Dir(path)
	parentInfo, err := os.Lstat(parent)
	if err != nil {
		return nil, fmt.Errorf("inspect owner API socket directory: %w", err)
	}
	stat, ok := parentInfo.Sys().(*syscall.Stat_t)
	if parentInfo.Mode()&os.ModeSymlink != 0 || !parentInfo.IsDir() || !ok || int(stat.Uid) != os.Getuid() || parentInfo.Mode().Perm()&0007 != 0 || parentInfo.Mode().Perm()&0020 != 0 {
		return nil, fmt.Errorf("owner API socket directory must be service-owned, non-world-accessible and not group-writable")
	}
	if existing, statErr := os.Lstat(path); statErr == nil {
		existingStat, validStat := existing.Sys().(*syscall.Stat_t)
		if existing.Mode()&os.ModeSymlink != 0 || existing.Mode()&os.ModeSocket == 0 || !validStat || int(existingStat.Uid) != os.Getuid() {
			return nil, fmt.Errorf("owner API socket path exists and is not a stale service socket")
		}
		connection, dialErr := net.DialTimeout("unix", path, 150*time.Millisecond)
		if dialErr == nil {
			_ = connection.Close()
			return nil, fmt.Errorf("owner API socket is already active")
		}
		if !errors.Is(dialErr, syscall.ECONNREFUSED) && !errors.Is(dialErr, os.ErrNotExist) {
			return nil, fmt.Errorf("owner API socket cannot be safely replaced: %w", dialErr)
		}
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("remove stale owner API socket: %w", err)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect owner API socket: %w", statErr)
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, fmt.Errorf("listen on owner API socket: %w", err)
	}
	listener.SetUnlinkOnClose(false)
	if err := os.Chmod(path, 0660); err != nil {
		_ = listener.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("secure owner API socket: %w", err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		_ = listener.Close()
		_ = os.Remove(path)
		return nil, err
	}
	return &OwnerSocketListener{UnixListener: listener, path: path, info: info}, nil
}

func (l *OwnerSocketListener) Close() error {
	err := l.UnixListener.Close()
	if current, statErr := os.Lstat(l.path); statErr == nil && os.SameFile(current, l.info) {
		if removeErr := os.Remove(l.path); err == nil {
			err = removeErr
		}
	}
	return err
}
