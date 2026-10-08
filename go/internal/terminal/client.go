// Package terminal discovers Herdr host terminals and restores their colors.
package terminal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrUnavailable = errors.New("owning host terminal unavailable")

const maxPayloadBytes = 1 << 20

// Client identifies a process and its terminal. PID keeps a reattach on the
// same Unix TTY distinct from the previous client.
type Client struct {
	PID         int
	Name, TTY   string
	Interactive bool // a UI client launch, local or --remote
	Remote      bool // a UI client rendering a server on another machine
	Bridge      bool // this machine's endpoint for another machine's --remote client
}

func IsHerdr(name string) bool {
	name = strings.ReplaceAll(name, "\\", "/")
	parts := strings.Split(name, "/")
	base := parts[len(parts)-1]
	return base == "herdr" || strings.EqualFold(base, "herdr.exe")
}

// IsInteractiveCommand recognizes client launches, including the executable
// in args[0]. CLI commands such as server stop, status, and plugin actions do
// not consume host color replies and must never receive queries. Unknown or
// unreadable arguments are rejected rather than guessed from the binary name.
func IsInteractiveCommand(args []string) bool {
	if len(args) == 0 || !IsHerdr(args[0]) {
		return false
	}
	args = args[1:]
	// Herdr rewrites this documented alias internally, without replacing argv.
	if len(args) == 3 && args[0] == "session" && args[1] == "attach" {
		name := args[2]
		return name != "" && name != "help" && !strings.HasPrefix(name, "-")
	}
	if len(args) > 0 && args[0] == "client" {
		args = args[1:]
	}
	for i := 0; i < len(args); i++ {
		flag, value, inline := strings.Cut(args[i], "=")
		switch flag {
		case "--handoff":
			if inline {
				return false
			}
		case "--session", "--remote", "--remote-keybindings":
			if !inline {
				i++
				if i >= len(args) {
					return false
				}
				value = args[i]
			}
			if value == "" || strings.HasPrefix(value, "-") {
				return false
			}
			if flag == "--remote-keybindings" && value != "local" && value != "server" {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// IsRemoteLaunch reports a UI client attached to another machine's server.
// Its theme still comes from this machine's config, like any local client.
func IsRemoteLaunch(args []string) bool {
	return IsInteractiveCommand(args) && hasRemoteFlag(args)
}

func hasRemoteFlag(args []string) bool {
	for _, arg := range args[1:] {
		if arg == "--remote" || strings.HasPrefix(arg, "--remote=") {
			return true
		}
	}
	return false
}

// IsRemoteBridge recognizes the process sshd starts on the server for a
// client that ran herdr --remote elsewhere. It has no local terminal.
func IsRemoteBridge(args []string) bool {
	return len(args) > 1 && IsHerdr(args[0]) && args[1] == "remote-client-bridge"
}

// classify is called only for processes whose name is Herdr.
func classify(p *Client, args []string) {
	p.Interactive = IsInteractiveCommand(args)
	p.Remote = p.Interactive && hasRemoteFlag(args)
	p.Bridge = IsRemoteBridge(args)
}

// OnlyRemoteClients reports that every client attached here came from another
// machine, so whoever opened a plugin popup cannot see this machine's theme.
// With local clients present the invoker is unknown and false is returned.
func OnlyRemoteClients(processes map[int]Client) bool {
	bridges := 0
	for _, p := range processes {
		if p.isInteractive() && !p.Remote {
			return false
		}
		if p.Bridge {
			bridges++
		}
	}
	return bridges > 0
}

func (c Client) isInteractive() bool {
	return IsHerdr(c.Name) && c.TTY != "" && c.Interactive
}

// SelectClient returns the explicitly requested client. It never falls back
// to another process by name.
func SelectClient(processes map[int]Client, explicit string) (Client, error) {
	pid, err := strconv.Atoi(explicit)
	if err != nil || pid <= 0 {
		return Client{}, fmt.Errorf("invalid HERDR_THEME_CLIENT_PID %q", explicit)
	}
	p, ok := processes[pid]
	if !ok || !p.isInteractive() {
		return Client{}, fmt.Errorf("PID %d is not a Herdr client", pid)
	}
	return p, nil
}

// ActiveClients returns the interactive Herdr clients on this machine, or only
// the explicit target. Every one of them reads this machine's config, so they
// share the selected theme. Remote bridges and CLI commands are never included.
func ActiveClients(explicit string) ([]Client, error) {
	processes, err := Processes()
	if err != nil {
		return nil, err
	}
	if explicit != "" {
		client, err := SelectClient(processes, explicit)
		if err != nil {
			return nil, err
		}
		return []Client{client}, nil
	}
	var clients []Client
	for _, p := range processes {
		if p.isInteractive() {
			clients = append(clients, p)
		}
	}
	if len(clients) == 0 {
		return nil, ErrUnavailable
	}
	return clients, nil
}

// parsePS reads `ps -o pid,tty,comm` output. Processes without a controlling
// terminal report "?" on Linux and "??" on macOS/BSD (#5); both become an
// empty TTY so the herdr server daemon is never mistaken for a client.
func parsePS(out string) map[int]Client {
	processes := make(map[int]Client)
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		tty := fields[1]
		if tty == "?" || tty == "??" || tty == "-" {
			tty = ""
		}
		processes[pid] = Client{PID: pid, TTY: tty, Name: strings.Join(fields[2:], " ")}
	}
	return processes
}
