package terminal

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// macOS kern.procargs2 returns argc, an executable path, alignment padding,
// then NUL-separated argv and environment strings. Only argv identifies the
// launch mode; spaces in a session name or executable path remain intact.
func darwinArgs(data []byte) ([]string, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("missing process argument count")
	}
	argc := int(binary.NativeEndian.Uint32(data[:4]))
	data = data[4:]
	end := bytes.IndexByte(data, 0)
	if argc <= 0 || argc > len(data) || end < 0 {
		return nil, fmt.Errorf("invalid process arguments")
	}
	data = bytes.TrimLeft(data[end+1:], "\x00")
	args := make([]string, 0, argc)
	for i := 0; i < argc; i++ {
		end := bytes.IndexByte(data, 0)
		if end < 0 {
			return nil, fmt.Errorf("truncated process arguments")
		}
		args = append(args, string(data[:end]))
		data = data[end+1:]
	}
	return args, nil
}
