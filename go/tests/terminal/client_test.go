package terminal_test

import (
	"testing"

	. "herdr-theme-picker/internal/terminal"
)

func TestTerminalClientSelection(t *testing.T) {
	processes := map[int]Client{
		10: {PID: 10, Name: "herdr.exe", TTY: "console", Interactive: true},
		20: {PID: 20, Name: "sh"},
		30: {PID: 30, Name: "/usr/bin/herdr", TTY: "pts/9", Interactive: true},
		40: {PID: 40, Name: "herdr"},
		60: {PID: 60, Name: "herdr.exe", TTY: "console"}, // server stop
	}
	client, err := SelectClient(processes, "30")
	if err != nil || client.PID != 30 {
		t.Fatal("explicit target ignored", err)
	}
	for _, pid := range []string{"", "invalid", "0", "-1", "20", "40", "60", "999"} {
		if _, err := SelectClient(processes, pid); err == nil {
			t.Errorf("accepted PID %q", pid)
		}
	}
}

func TestInteractiveLaunchModes(t *testing.T) {
	for _, args := range [][]string{
		{"herdr"},
		{"/usr/local/bin/herdr", "client"},
		{`C:\Program Files\Herdr\herdr.exe`, "--session", "server"},
		{"herdr", "--session=work projects"},
		{"herdr", "--remote", "dev-host", "--remote-keybindings", "server"},
		{"herdr", "--remote=dev-host", "--handoff"},
	} {
		if !IsInteractiveCommand(args) {
			t.Errorf("rejected interactive launch: %q", args)
		}
	}
	for _, args := range [][]string{
		nil, {}, {""}, {"pwsh"},
		{"herdr", "server"},
		{"herdr", "server", "stop"},
		{"herdr", "server", "reload-config"},
		{"herdr", "--session", "work", "server", "stop"},
		{"herdr", "status"},
		{"herdr", "plugin", "action", "invoke", "sync"},
		{"herdr", "pane", "list"},
		{"herdr", "update"},
		{"herdr", "remote-client-bridge"},
		{"herdr", "--machine", "dev", "status"},
		{"herdr", "--help"},
		{"herdr", "--version"},
		{"herdr", "--session"},
		{"herdr", "--session="},
		{"herdr", "--session", "--help"},
		{"herdr", "--remote-keybindings", "unknown"},
		{"herdr", "--unknown"},
	} {
		if IsInteractiveCommand(args) {
			t.Errorf("accepted CLI or unknown launch: %q", args)
		}
	}
}

func TestInvalidExplicitTargetNeverFallsBack(t *testing.T) {
	for _, pid := range []string{"invalid", "0", "-1"} {
		if clients, err := ActiveClients(pid); err == nil || len(clients) != 0 {
			t.Fatalf("invalid target %q selected clients: %v %v", pid, clients, err)
		}
	}
}

func TestRemoteLaunchClassification(t *testing.T) {
	for _, args := range [][]string{
		{"herdr", "--remote", "mac"},
		{`C:\Herdr\herdr.exe`, "--remote=mac", "--session", "work"},
	} {
		if !IsRemoteLaunch(args) || IsRemoteBridge(args) {
			t.Errorf("not a remote UI client: %q", args)
		}
	}
	for _, args := range [][]string{{"herdr"}, {"herdr", "--session", "remote"}, {"herdr", "session", "attach", "remote"}} {
		if IsRemoteLaunch(args) {
			t.Errorf("local client treated as remote: %q", args)
		}
	}
	for _, args := range [][]string{
		{"/opt/homebrew/bin/herdr", "remote-client-bridge", "--idle-timeout-v1"},
		{"herdr", "remote-client-bridge"},
	} {
		if !IsRemoteBridge(args) || IsInteractiveCommand(args) {
			t.Errorf("bridge misclassified: %q", args)
		}
	}
	for _, args := range [][]string{{"herdr"}, {"herdr", "server"}, {"sshd", "remote-client-bridge"}} {
		if IsRemoteBridge(args) {
			t.Errorf("not a bridge: %q", args)
		}
	}
}

// The picker popup shows client-setup instructions only when every attached
// client is on another machine; with any local client it cannot know who
// opened it and keeps the picker.
func TestOnlyRemoteClients(t *testing.T) {
	server := Client{PID: 1, Name: "/opt/homebrew/bin/herdr"}
	local := Client{PID: 2, Name: "herdr", TTY: "ttys001", Interactive: true}
	bridge := Client{PID: 3, Name: "/opt/homebrew/bin/herdr", Bridge: true}
	outbound := Client{PID: 4, Name: "herdr", TTY: "ttys002", Interactive: true, Remote: true}
	shell := Client{PID: 5, Name: "zsh", TTY: "ttys003"}
	for name, tc := range map[string]struct {
		processes []Client
		want      bool
	}{
		"nothing attached":           {[]Client{server, shell}, false},
		"local only":                 {[]Client{server, local}, false},
		"remote only":                {[]Client{server, bridge, shell}, true},
		"local and remote":           {[]Client{server, local, bridge}, false},
		"outbound remote and bridge": {[]Client{server, outbound, bridge}, true},
		"outbound remote only":       {[]Client{server, outbound}, false},
	} {
		processes := make(map[int]Client)
		for _, p := range tc.processes {
			processes[p.PID] = p
		}
		if got := OnlyRemoteClients(processes); got != tc.want {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
		}
	}
}

func TestSessionAttachLaunch(t *testing.T) {
	for _, name := range []string{"work", "side-project", "default"} {
		if !IsInteractiveCommand([]string{"herdr", "session", "attach", name}) {
			t.Fatalf("rejected named session %s", name)
		}
	}
	for _, args := range [][]string{
		{"herdr", "session", "attach"}, {"herdr", "session", "attach", "--help"},
		{"herdr", "session", "attach", "help"}, {"herdr", "session", "attach", "work", "--help"},
		{"herdr", "session", "stop", "work"},
	} {
		if IsInteractiveCommand(args) {
			t.Fatalf("accepted CLI launch: %q", args)
		}
	}
}
