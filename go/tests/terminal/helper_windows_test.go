//go:build windows

package terminal_test

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"

	"herdr-theme-picker/internal/terminal"
	"herdr-theme-picker/tests/testutil"
)

func TestMain(m *testing.M) { testutil.Main(m) }

func TestExactHelperDoesNotFallBackToAnotherHerdr(t *testing.T) {
	// The test executable is not Herdr. Even if real clients are running, a
	// request for this PID must fail without attaching to any of them.
	// An override also must not redirect a daemon's exact target.
	t.Setenv("HERDR_THEME_CLIENT_PID", "invalid")
	err := terminal.EmitClient(terminal.Client{PID: os.Getpid()}, "")
	if !errors.Is(err, terminal.ErrUnavailable) {
		t.Fatalf("wrong target did not report unavailable: %v", err)
	}
}

func TestInvalidExplicitPIDIsRejectedBeforeHelper(t *testing.T) {
	if _, err := terminal.ActiveClients("invalid"); err == nil || errors.Is(err, terminal.ErrUnavailable) {
		t.Fatalf("invalid override was ignored: %v", err)
	}
}

func TestHelperWritesToAnIsolatedClientConsole(t *testing.T) {
	pid := startIsolatedHerdr(t)
	t.Setenv("HERDR_THEME_CLIENT_PID", "invalid")
	if err := terminal.EmitClient(terminal.Client{PID: pid}, "\033]11;#123456\007"); err != nil {
		t.Fatalf("could not write to isolated client console: %v", err)
	}
}

func TestCLICommandsNeverReceiveHostQueries(t *testing.T) {
	for _, args := range [][]string{
		{"server", "stop"}, {"server", "reload-config"},
		{"plugin", "action", "invoke", "sync"}, {"status"}, {"--help"},
	} {
		t.Run(args[0]+"-"+args[len(args)-1], func(t *testing.T) {
			pid := startIsolatedHerdr(t, args...)
			// A CLI process has a real console and the same executable name.
			// Neither discovery nor an explicit PID can authorize recoloring it.
			clients, err := terminal.ActiveClients("")
			if err != nil && !errors.Is(err, terminal.ErrUnavailable) {
				t.Fatal(err)
			}
			for _, c := range clients {
				if c.PID == pid {
					t.Fatal("CLI command was discovered as an attached client")
				}
			}
			if _, err := terminal.ActiveClients(strconv.Itoa(pid)); err == nil {
				t.Fatal("explicit PID bypassed CLI rejection")
			}
			if err := terminal.EmitClient(terminal.Client{PID: pid}, "\033]11;#123456\007"); !errors.Is(err, terminal.ErrUnavailable) {
				t.Fatalf("helper accepted CLI command: %v", err)
			}
		})
	}
}

func TestClientDiscoveryPreservesSessionArguments(t *testing.T) {
	pid := startIsolatedHerdr(t, "--session", "work projects")
	clients, err := terminal.ActiveClients(strconv.Itoa(pid))
	if err != nil || len(clients) != 1 || clients[0].PID != pid {
		t.Fatalf("could not discover named-session client: %v %v", clients, err)
	}
}

func startIsolatedHerdr(t *testing.T, args ...string) int {
	t.Helper()
	// A copied test executable named herdr.exe owns a hidden console. This tests
	// AttachConsole and CONOUT$ access without touching the user's terminal.
	path := filepath.Join(t.TempDir(), "herdr.exe")
	testutil.CopyExecutable(t, path)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	client := exec.CommandContext(ctx, path, args...)
	client.Env = append(os.Environ(), "THEME_TEST_CONSOLE_CLIENT=1")
	client.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NEW_CONSOLE}
	stdin, err := client.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := client.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stdin.Close()
		cancel()
		client.Wait()
	})
	ready, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || ready != "ready\n" {
		t.Fatalf("test client did not start: %q %v", ready, err)
	}
	return client.Process.Pid
}
