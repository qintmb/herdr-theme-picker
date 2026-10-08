//go:build !windows && !linux && !darwin

package terminal

import "fmt"

func processArgs(int) ([]string, error) {
	return nil, fmt.Errorf("process arguments unavailable on this platform")
}
