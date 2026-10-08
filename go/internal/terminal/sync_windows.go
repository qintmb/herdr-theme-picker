//go:build windows

package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	consoleDLL    = windows.NewLazySystemDLL("kernel32.dll")
	freeConsole   = consoleDLL.NewProc("FreeConsole")
	attachConsole = consoleDLL.NewProc("AttachConsole")
)

func Processes() (map[int]Client, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, err
	}
	processes := make(map[int]Client)
	for {
		pid := int(entry.ProcessID)
		p := Client{PID: pid, Name: windows.UTF16ToString(entry.ExeFile[:]), TTY: "console"}
		if IsHerdr(p.Name) {
			if args, err := processArgs(pid); err == nil {
				classify(&p, args)
			}
		}
		processes[pid] = p
		err := windows.Process32Next(snapshot, &entry)
		if err == windows.ERROR_NO_MORE_FILES {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return processes, nil
}

// EmitClient always writes to the requested client. Environment overrides and
// unrelated processes must not turn a failed write into a false success.
func EmitClient(client Client, payload string) error {
	return runHelper(HostPayload(payload), strconv.Itoa(client.PID))
}

func runHelper(payload string, args ...string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, append([]string{"terminal-sync"}, args...)...)
	cmd.Stdin = strings.NewReader(payload)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
	output, err := cmd.CombinedOutput()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 3 {
			return ErrUnavailable
		}
		return fmt.Errorf("sync terminal: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// RunHelper must be called only by the short-lived terminal-sync subprocess.
// It attaches to exactly the requested client and never falls back.
func RunHelper(pidText string) error {
	pid, err := strconv.Atoi(pidText)
	if err != nil || pid <= 0 {
		return fmt.Errorf("invalid client PID %q", pidText)
	}
	data, err := io.ReadAll(io.LimitReader(os.Stdin, maxPayloadBytes+1))
	if err != nil {
		return err
	}
	if len(data) > maxPayloadBytes {
		return fmt.Errorf("terminal payload is too large")
	}
	processes, err := Processes()
	if err != nil {
		return err
	}
	if ret, _, err := freeConsole.Call(); ret == 0 {
		return fmt.Errorf("detach helper console: %w", err)
	}
	p, ok := processes[pid]
	if !ok || !p.isInteractive() {
		return ErrUnavailable
	}
	if ret, _, attachErr := attachConsole.Call(uintptr(pid)); ret == 0 {
		if errors.Is(attachErr, windows.ERROR_INVALID_HANDLE) || errors.Is(attachErr, windows.ERROR_INVALID_PARAMETER) {
			return ErrUnavailable
		}
		return fmt.Errorf("attach client console: %w", attachErr)
	}
	defer freeConsole.Call()
	f, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	var mode uint32
	if err := windows.GetConsoleMode(windows.Handle(f.Fd()), &mode); err != nil {
		return err
	}
	if err := windows.SetConsoleMode(windows.Handle(f.Fd()), mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return err
	}
	defer windows.SetConsoleMode(windows.Handle(f.Fd()), mode)
	_, err = f.Write(data)
	return err
}
