package terminal

import (
	"fmt"
	"strings"
)

// HostPayload sets colors, then requests their current values. Herdr consumes
// OSC replies as host-theme updates and propagates them to terminal runtimes.
// Without readback, a reattached client can keep rendering the colors it queried
// before the daemon restored the host palette.
func HostPayload(colors string) string {
	if colors == "" {
		return ""
	}
	var out strings.Builder
	out.WriteString(colors)
	out.WriteString("\033]10;?\033\\\033]11;?\033\\")
	for index := 0; index < 16; index++ {
		fmt.Fprintf(&out, "\033]4;%d;?\033\\", index)
	}
	return out.String()
}
