//go:build !windows

package theme

import "os"

func replaceFile(from, to string) error { return os.Rename(from, to) }
