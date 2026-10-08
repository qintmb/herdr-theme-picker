// Package instance provides an OS-held lock for a background process.
package instance

import (
	"errors"
	"os"
)

var ErrLocked = errors.New("another instance holds the lock")

// Acquire holds path's advisory lock until the returned file is closed.
// The lock file must stay on disk so every process locks the same file.
func Acquire(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
