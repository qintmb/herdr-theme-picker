package terminal

import (
	"bytes"
	"os"
	"strconv"
)

func processArgs(pid int) ([]string, error) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		return nil, err
	}
	var args []string
	for _, arg := range bytes.Split(bytes.TrimSuffix(data, []byte{0}), []byte{0}) {
		args = append(args, string(arg))
	}
	return args, nil
}
