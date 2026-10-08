package terminal

import "golang.org/x/sys/unix"

func processArgs(pid int) ([]string, error) {
	data, err := unix.SysctlRaw("kern.procargs2", pid)
	if err != nil {
		return nil, err
	}
	return darwinArgs(data)
}
