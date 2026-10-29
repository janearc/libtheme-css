package smoke_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	// succeeded is the status of a command that did what it was asked. It
	// is defined by the shell.
	succeeded = 0
	// failed is the status of one that did not. It is defined by the shell.
	failed = 1
)

// exitCode is the status a finished command returned, and its output.
func exitCode(t *testing.T, binary string, arguments ...string) (int, string) {
	t.Helper()
	output, err := exec.Command(binary, arguments...).CombinedOutput()
	var exited *exec.ExitError
	switch {
	case err == nil:
		return succeeded, string(output)
	case errors.As(err, &exited):
		return exited.ExitCode(), string(output)
	}
	t.Fatalf("%v: %v", arguments, err)
	return succeeded, ""
}

// The smoke test: the built command exits 0 for a source inside the
// reader's range and 1, naming the colour, for one outside it.
func TestCheckExitsWithTheVerdict(t *testing.T) {
	binary := built(t)
	dir := t.TempDir()
	inside, outside := filepath.Join(dir, "inside.css"), filepath.Join(dir, "outside.css")
	band := strings.ReplaceAll(duskSheet, "%s", "#c8c8c8")
	if err := os.WriteFile(inside, []byte(inRange), private); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte(band), private); err != nil {
		t.Fatal(err)
	}
	if code, output := exitCode(t, binary, "check", inside); code != succeeded {
		t.Errorf("a source in range exited %d:\n%s", code, output)
	}
	code, output := exitCode(t, binary, "check", outside)
	if code != failed || !strings.Contains(output, "too dim to read at length") {
		t.Errorf("a source out of range exited %d:\n%s", code, output)
	}
	if code, _ := exitCode(t, binary, "check"); code == succeeded {
		t.Error("check with nothing to check exited 0")
	}
}

// inRange is a source inside the reference range on every surface.
const inRange = `:root {
  --ground: #0a0a0a; --cursor-line: #101010; --surface-1: #161616;
  --panel: #0e0e0e;
  --ink: #c8c8c8; --dim: #c4c4c4; --heading: #c6c6c6; --link: #c5c5c5;
  --code: #c7c7c7; --string: #c9c9c9; --gutter: #4a4a4a; --border: #3a3a3a;
  --err: #e07070; --warn: #d9a050; --info: #70a0e0; --hint: #70d0b0;
  --black: #101010; --red: #aa5555; --green: #55aa55; --yellow: #aaaa55;
  --blue: #5555aa; --magenta: #aa55aa; --cyan: #55aaaa; --white: #aaaaaa;
  --bright-black: #555555; --bright-red: #ff5555; --bright-green: #55ff55;
  --bright-yellow: #ffff55; --bright-blue: #5555ff;
  --bright-magenta: #ff55ff; --bright-cyan: #55ffff; --bright-white: #ffffff;
}
`
