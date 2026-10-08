package terminal

import (
	"encoding/binary"
	"reflect"
	"testing"
)

func darwinFixture(args []string, environment string) []byte {
	data := make([]byte, 4)
	binary.NativeEndian.PutUint32(data, uint32(len(args)))
	data = append(data, []byte("/Applications/Herdr App/herdr\x00\x00\x00")...)
	for _, arg := range args {
		data = append(data, []byte(arg)...)
		data = append(data, 0)
	}
	return append(data, []byte(environment)...)
}

func TestDarwinArgumentsPreserveBoundariesAndExcludeEnvironment(t *testing.T) {
	for _, args := range [][]string{
		{"/Applications/Herdr App/herdr", "--session", "work projects"},
		{"herdr", "server", "stop"},
		{"herdr", "--session", ""},
	} {
		got, err := darwinArgs(darwinFixture(args, "MODE=client\x00"))
		if err != nil || !reflect.DeepEqual(got, args) {
			t.Fatalf("wrong macOS argv: %q %v; want %q", got, err, args)
		}
		if IsInteractiveCommand(got) != IsInteractiveCommand(args) {
			t.Fatal("decoding changed launch mode")
		}
	}
}

func TestDarwinArgumentsRejectMalformedBuffers(t *testing.T) {
	truncated := darwinFixture([]string{"herdr", "server", "stop"}, "")
	for _, data := range [][]byte{
		nil, {1, 0, 0}, {0, 0, 0, 0},
		truncated[:len(truncated)-1],
		{255, 255, 255, 255, 'h', 0},
	} {
		if _, err := darwinArgs(data); err == nil {
			t.Errorf("accepted malformed argv: %v", data)
		}
	}
}
