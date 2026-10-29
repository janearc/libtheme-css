package smoke_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sourceAt writes text to a file called base in a folder of its own, and
// returns the path.
func sourceAt(t *testing.T, base, sheet string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), base)
	if err := os.WriteFile(path, []byte(sheet), private); err != nil {
		t.Fatal(err)
	}
	return path
}

// readOnly is the permission of a folder nothing can be written into. It is
// defined by the file mode bits.
const readOnly = os.FileMode(0o500)

// usageStatus is the exit status of a command called wrongly. It is
// defined by the shell's convention for it.
const usageStatus = 2

// The anti-tests: a render that could write nothing says so and exits 1,
// whether the place for a folder is a file, or the destination cannot be
// written to. It used to call every failure "skipped" and exit 0.
func TestRenderThatWritesNothingFails(t *testing.T) {
	binary := built(t)
	source := sourceAt(t, "dusk.css", strings.ReplaceAll(duskSheet, "%s", "#c8c8c8"))
	blocked := t.TempDir()
	for _, program := range []string{"ghostty", "nvim", "vim", "glamour"} {
		if err := os.WriteFile(filepath.Join(blocked, program), nil, private); err != nil {
			t.Fatal(err)
		}
	}
	code, output := exitCode(t, binary, "render", source, blocked)
	if code != failed || !strings.Contains(output, "could not be written") {
		t.Errorf("render into a place of files exited %d:\n%s", code, output)
	}
	sealed := t.TempDir()
	if err := os.Chmod(sealed, readOnly); err != nil {
		t.Fatal(err)
	}
	code, output = exitCode(t, binary, "render", source, sealed)
	if code != failed || !strings.Contains(output, "could not be written") {
		t.Errorf("render into a sealed folder exited %d:\n%s", code, output)
	}
}

// A render that writes some files and cannot write others exits 1 too, and
// the files it could write are there.
func TestRenderThatWritesSomeFails(t *testing.T) {
	binary := built(t)
	source := sourceAt(t, "dusk.css", strings.ReplaceAll(duskSheet, "%s", "#c8c8c8"))
	partial := t.TempDir()
	if err := os.WriteFile(filepath.Join(partial, "glamour"), nil, private); err != nil {
		t.Fatal(err)
	}
	code, output := exitCode(t, binary, "render", source, partial)
	if code != failed {
		t.Errorf("a half-written render exited %d:\n%s", code, output)
	}
	if _, err := os.Stat(filepath.Join(partial, "ghostty", "dusk")); err != nil {
		t.Error("the files that could be written were not")
	}
}

// A render that writes everything exits 0, and a source that fills no
// program at all exits 1 saying so: there is nothing to show for it.
func TestRenderSaysWhenThereIsNothingToWrite(t *testing.T) {
	binary := built(t)
	good := strings.ReplaceAll(duskSheet, "%s", "#c8c8c8")
	if code, output := exitCode(t, binary, "render", sourceAt(t, "dusk.css", good), t.TempDir()); code != succeeded {
		t.Errorf("a whole render exited %d:\n%s", code, output)
	}
	bare := sourceAt(t, "bare.css", ":root { --ground: #101010; --ink: #c8c8c8; }\n")
	code, output := exitCode(t, binary, "render", bare, t.TempDir())
	if code != failed || !strings.Contains(output, "nothing could be rendered") {
		t.Errorf("a source that fills nothing exited %d:\n%s", code, output)
	}
}

// The anti-tests: an option nobody knows is a wrong call, exit 2, not an
// option ignored.
func TestUnknownOptionsAreWrongCalls(t *testing.T) {
	binary := built(t)
	theme := sourceAt(t, "theme", "foreground = #ffffff\n")
	for name, arguments := range map[string][]string{
		"sheet alone":       {"sheet", "--bogus", theme},
		"sheet beside":      {"sheet", "--ghostty", theme, "--bogus", theme},
		"sheet odd":         {"sheet", "--ghostty"},
		"sheet nothing":     {"sheet"},
		"diff":              {"diff", "--bogus", theme, theme},
		"diff git":          {"diff", "--git", "--bogus"},
		"render few":        {"render", theme},
		"show none":         {"show"},
		"no verb":           {},
		"a verb nobody has": {"bogus"},
	} {
		if code, output := exitCode(t, binary, arguments...); code != usageStatus {
			t.Errorf("%s exited %d:\n%s", name, code, output)
		}
	}
}

// Calls that are right still work: sheet of a terminal theme prints a
// source, diff of a file with itself exits 0, and show prints roles.
func TestRightCallsStillWork(t *testing.T) {
	binary := built(t)
	theme := sourceAt(t, "theme", "foreground = #ffffff\nbackground = #000000\n")
	for name, arguments := range map[string][]string{
		"sheet": {"sheet", "--ghostty", theme},
		"same":  {"diff", theme, theme},
		"show":  {"show", theme},
		"age":   {"--age"},
	} {
		if code, output := exitCode(t, binary, arguments...); code != succeeded {
			t.Errorf("%s exited %d:\n%s", name, code, output)
		}
	}
}
