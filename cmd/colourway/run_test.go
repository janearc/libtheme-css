package main

import (
	"bytes"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/internal/host"
	"github.com/janearc/libtheme-css/internal/mocks"
	"github.com/janearc/libtheme-css/spaces/srgb"
	"go.uber.org/mock/gomock"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sourceCSS is a source with every colour the programs need.
const sourceCSS = `:root {
  --ground: #202020; --ink: #c8c8c8; --heading: #aa5555;
  --black: #101010; --red: #aa5555; --green: #55aa55; --yellow: #aaaa55;
  --blue: #5555aa; --magenta: #aa55aa; --cyan: #55aaaa; --white: #aaaaaa;
  --bright-black: #555555; --bright-red: #ff5555; --bright-green: #55ff55;
  --bright-yellow: #ffff55; --bright-blue: #5555ff;
  --bright-magenta: #ff55ff; --bright-cyan: #55ffff; --bright-white: #ffffff;
}
`

// ran is a call of the command, with what it printed to each stream.
type ran struct {
	status    int
	out, errs string
}

// called runs the command with the arguments.
func called(arguments ...string) ran {
	var out, errs bytes.Buffer
	status := run(arguments, streams{out: &out, errs: &errs})
	return ran{status, out.String(), errs.String()}
}

// Every way of calling the command wrongly is a wrong call: it says what
// was wrong, shows the usage, and exits 2.
func TestWrongCallsExitTwo(t *testing.T) {
	theme := stored(t, "theme", "foreground = #ffffff\n")
	for name, arguments := range map[string][]string{
		"no verb":         {},
		"unknown verb":    {"bogus"},
		"render few":      {"render", theme},
		"render many":     {"render", theme, theme, theme},
		"show none":       {"show"},
		"show many":       {"show", theme, theme},
		"sheet none":      {"sheet"},
		"sheet bogus":     {"sheet", "--bogus", theme},
		"sheet odd":       {"sheet", "--ghostty"},
		"diff bogus":      {"diff", "--bogus", theme, theme},
		"diff one":        {"diff", theme},
		"diff git bogus":  {"diff", "--git", "--bogus"},
		"check no source": {"check", "--policy"},
	} {
		got := called(arguments...)
		if name == "check no source" {
			if got.status != failed {
				t.Errorf("%s exited %d", name, got.status)
			}
			continue
		}
		if got.status != wrongCall || !strings.Contains(got.errs, "usage:") {
			t.Errorf("%s exited %d: %s", name, got.status, got.errs)
		}
	}
}

// The right calls work: the age, a render of a whole source, a diff of a
// file with itself and with another, show, and sheet.
func TestRightCallsWork(t *testing.T) {
	source := stored(t, "dusk.css", sourceCSS)
	other := stored(t, "dusk.css", strings.ReplaceAll(sourceCSS, "#c8c8c8", "#d8d8d8"))
	theme := stored(t, "theme", "foreground = #ffffff\nbackground = #000000\n")
	folder := t.TempDir()
	cases := []struct {
		arguments []string
		status    int
		want      string
	}{
		{[]string{"--age"}, succeeded, "colourway"},
		{[]string{"version"}, succeeded, "colourway"},
		{[]string{"render", source, folder}, succeeded, "ghostty"},
		{[]string{"diff", source, source}, succeeded, "0 changed"},
		{[]string{"diff", source, other}, failed, "1 changed"},
		{[]string{"diff", "--colour", source, other}, failed, "1 changed"},
		{[]string{"show", source}, succeeded, "ink"},
		{[]string{"sheet", "--ghostty", theme}, succeeded, "--ink"},
		{[]string{"show", "/no/such/file"}, failed, ""},
		{[]string{"check", source}, failed, ""},
	}
	for _, one := range cases {
		got := called(one.arguments...)
		if got.status != one.status || !strings.Contains(got.out, one.want) {
			t.Errorf("%v exited %d, printed %q, said %q", one.arguments, got.status,
				got.out, got.errs)
		}
	}
}

// diff says nothing on stderr when the files merely differ: that is its
// answer, not an error.
func TestDifferingIsNotAnError(t *testing.T) {
	source := stored(t, "dusk.css", sourceCSS)
	other := stored(t, "dusk.css", strings.ReplaceAll(sourceCSS, "#c8c8c8", "#d8d8d8"))
	if got := called("diff", source, other); got.errs != "" {
		t.Errorf("a diff of differing files said %q", got.errs)
	}
}

// Render tells the three apart: written, skipped (the source cannot fill
// the program) and failed (it could not be written), and exits by them.
func TestRenderTellsSkipFromFailure(t *testing.T) {
	bare := stored(t, "bare.css", ":root { --ground: #101010; --ink: #c8c8c8; }\n")
	got := called("render", bare, t.TempDir())
	if got.status != failed || !strings.Contains(got.out, "skipped") ||
		!strings.Contains(got.errs, "nothing could be rendered") {
		t.Errorf("a bare source: %d %q %q", got.status, got.out, got.errs)
	}
	source := stored(t, "dusk.css", sourceCSS)
	blocked := t.TempDir()
	if err := os.WriteFile(filepath.Join(blocked, "glamour"), nil, private); err != nil {
		t.Fatal(err)
	}
	got = called("render", source, blocked)
	if got.status != failed || !strings.Contains(got.out, "failed") ||
		!strings.Contains(got.errs, "files could not be written") {
		t.Errorf("a blocked folder: %d %q %q", got.status, got.out, got.errs)
	}
	got = called("render", "/no/such/source.css", t.TempDir())
	if got.status != failed {
		t.Errorf("a missing source exited %d", got.status)
	}
}

// Sheet reads each program's file, says when two disagree, and fails when a
// file is not there or not readable.
func TestSheetReadsAndComplains(t *testing.T) {
	theme := stored(t, "theme", "foreground = #ffffff\nbackground = #000000\n")
	scheme := stored(t, "scheme.lua", "vim.g.colors_name = \"x\"\nlocal c = {\n  bg = \"#111111\",\n  ink = \"#eeeeee\",\n}\n")
	style := stored(t, "style.json", `{"document": {"color": "#dddddd"}}`)
	got := called("sheet", "--ghostty", theme, "--nvim", scheme, "--glamour", style)
	if got.status != succeeded || !strings.Contains(got.errs, "kept") {
		t.Errorf("a sheet of three: %d %q %q", got.status, got.out, got.errs)
	}
	for _, option := range []string{"--ghostty", "--nvim", "--glamour"} {
		if got := called("sheet", option, "/no/such/file"); got.status != failed {
			t.Errorf("sheet %s of a missing file exited %d", option, got.status)
		}
		if got := called("sheet", option, stored(t, "bad", "{{{ not a theme")); got.status != failed &&
			option != "--ghostty" {
			t.Errorf("sheet %s of rubbish exited %d", option, got.status)
		}
	}
}

// A terminal theme with a comment on top and a claude code colour in its
// palette is a source with the comment as its opening words and the colour
// named, and a role two files agree on is no complaint.
func TestSheetNamesClaudeColoursAndKeepsTheComment(t *testing.T) {
	theme := stored(t, "theme", "# a theme\n# in two lines\nforeground = #ffffff\n"+
		"background = #000000\npalette = 252=#c5e699\npalette = 1=#aa5555\n")
	scheme := stored(t, "scheme.lua", "vim.g.colors_name = \"x\"\nlocal c = {\n"+
		"  bg = \"#000000\",\n  ink = \"#ffffff\",\n}\n")
	got := called("sheet", "--ghostty", theme, "--nvim", scheme)
	if got.status != succeeded || got.errs != "" {
		t.Fatalf("a sheet of two that agree: %d %q", got.status, got.errs)
	}
	for _, want := range []string{"a theme\n * in two lines", "--claude-", "--red"} {
		if !strings.Contains(got.out, want) {
			t.Errorf("no %q in\n%s", want, got.out)
		}
	}
	if strings.Contains(got.out, "--palette-252") {
		t.Error("a claude code colour was left as its palette number")
	}
}

// A theme the terminal reader refuses is a failure, and a style that says
// another ink than the theme is told, and the first is kept.
func TestSheetFailsOnABadThemeAndKeepsTheFirst(t *testing.T) {
	bad := stored(t, "theme", "palette = 300=#000000\n")
	if got := called("sheet", "--ghostty", bad); got.status != failed {
		t.Errorf("a bad theme exited %d", got.status)
	}
	theme := stored(t, "theme", "foreground = #ffffff\nbackground = #000000\n")
	style := stored(t, "style.json", `{"document": {"color": "#dddddd"}}`)
	if got := called("sheet", "--ghostty", theme, "--glamour", style); got.status != succeeded {
		t.Errorf("a theme and a style: %d %q", got.status, got.errs)
	}
	into, from := css.New(), css.New()
	into.Set("ink", srgb.MustHex("#ffffff").Swatch())
	from.Set("ink", srgb.MustHex("#dddddd").Swatch())
	var told bytes.Buffer
	merge(&told, into, from, "style.json", "")
	if !strings.Contains(told.String(), "kept #ffffff") {
		t.Errorf("two inks, said %q", told.String())
	}
}

// process is the program main runs in, for a test: a mock to stop at, and the
// files its output went to.
type process struct {
	stopper   *mocks.MockStopper
	out, errs *os.File
}

// was is what a test changed about the program, and puts it back.
type was struct {
	arguments []string
	out, errs *os.File
	stopsWith host.Stopper
	made      process
}

// restore puts the program back as it was.
func (before was) restore() {
	os.Args, os.Stdout, os.Stderr, program = before.arguments, before.out,
		before.errs, before.stopsWith
	before.made.out.Close()
	before.made.errs.Close()
}

// emptyFile is a new empty file for a test's output.
func emptyFile(t *testing.T) *os.File {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "output-*")
	if err != nil {
		t.Fatal(err)
	}
	return file
}

// asProcess makes the test the program: main sees these arguments, writes
// to two empty files, and stops at the mock it is given. The program is put
// back when the test ends.
func asProcess(t *testing.T, arguments ...string) process {
	t.Helper()
	made := process{
		stopper: mocks.NewMockStopper(gomock.NewController(t)),
		out:     emptyFile(t),
		errs:    emptyFile(t),
	}
	before := was{os.Args, os.Stdout, os.Stderr, program, made}
	t.Cleanup(before.restore)
	os.Args = append([]string{"colourway"}, arguments...)
	os.Stdout, os.Stderr, program = made.out, made.errs, made.stopper
	return made
}

// printed is what main wrote to a file.
func printed(t *testing.T, file *os.File) string {
	t.Helper()
	raw, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// main runs the verb it is called with and stops the program with its
// status, and says nothing to the error stream.
func TestMainStopsWithTheStatusOfTheVerb(t *testing.T) {
	made := asProcess(t, "version")
	made.stopper.EXPECT().Exit(succeeded)
	main()
	if !strings.Contains(printed(t, made.out), "colourway") {
		t.Errorf("version said %q", printed(t, made.out))
	}
	if said := printed(t, made.errs); said != "" {
		t.Errorf("version complained: %q", said)
	}
}

// The anti-test: a call main cannot make sense of stops the program with
// the wrong-call status, and shows the usage on the error stream and not on
// the output.
func TestMainStopsWithTheWrongCallStatus(t *testing.T) {
	made := asProcess(t)
	made.stopper.EXPECT().Exit(wrongCall)
	main()
	if !strings.Contains(printed(t, made.errs), "usage") {
		t.Errorf("no usage on the error stream: %q", printed(t, made.errs))
	}
	if said := printed(t, made.out); said != "" {
		t.Errorf("the usage went to the output: %q", said)
	}
}
