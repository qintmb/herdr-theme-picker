package terminal

import "testing"

// Issue #5: Linux reports "?" for daemons without a controlling terminal and
// macOS reports "??". Neither may be treated as the herdr client's TTY.
func TestParsePSDropsDaemonTTYs(t *testing.T) {
	cases := map[string]struct {
		out    string
		server int
		client int
		tty    string
	}{
		"linux": {
			out:    "    PID TT       COMMAND\n    100 ?        herdr\n    200 pts/3    herdr\n",
			server: 100, client: 200, tty: "pts/3",
		},
		"macos": {
			out:    "  PID TTY      COMM\n  100 ??       /opt/homebrew/bin/herdr\n  200 ttys004  /opt/homebrew/bin/herdr\n",
			server: 100, client: 200, tty: "ttys004",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			processes := parsePS(tc.out)
			if len(processes) != 2 {
				t.Fatalf("parsed %d processes, want 2: %+v", len(processes), processes)
			}
			if tty := processes[tc.server].TTY; tty != "" {
				t.Fatalf("server TTY = %q, want empty", tty)
			}
			if tty := processes[tc.client].TTY; tty != tc.tty {
				t.Fatalf("client TTY = %q, want %q", tty, tc.tty)
			}
		})
	}
}
