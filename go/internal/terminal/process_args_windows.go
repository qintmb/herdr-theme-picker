package terminal

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Read arguments directly from the OS; do not launch a shell or guess a
// process's role from its executable name. Unreadable processes are skipped.
func processArgs(pid int) ([]string, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(handle)
	var size uint32
	err = windows.NtQueryInformationProcess(handle, windows.ProcessCommandLineInformation, nil, 0, &size)
	if err != nil && !errors.Is(err, windows.STATUS_INFO_LENGTH_MISMATCH) && !errors.Is(err, windows.STATUS_BUFFER_TOO_SMALL) {
		return nil, err
	}
	if size < uint32(unsafe.Sizeof(windows.NTUnicodeString{})) || size > maxPayloadBytes {
		return nil, fmt.Errorf("invalid process command line size")
	}
	buffer := make([]byte, size)
	if err := windows.NtQueryInformationProcess(handle, windows.ProcessCommandLineInformation, unsafe.Pointer(&buffer[0]), size, &size); err != nil {
		return nil, err
	}
	text := (*windows.NTUnicodeString)(unsafe.Pointer(&buffer[0]))
	offset := uintptr(unsafe.Pointer(text.Buffer)) - uintptr(unsafe.Pointer(&buffer[0]))
	if text.Length%2 != 0 || offset > uintptr(len(buffer)) || uintptr(text.Length) > uintptr(len(buffer))-offset {
		return nil, fmt.Errorf("invalid process command line buffer")
	}
	return windows.DecomposeCommandLine(text.String())
}
