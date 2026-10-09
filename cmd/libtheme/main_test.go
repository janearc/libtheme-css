package main

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/spaces/radiation"
)

// capture runs f with stdout caught, so a test can read what a verb
// printed.
func capture(t *testing.T, f func() error) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	if err := f(); err != nil {
		t.Fatal(err)
	}
	w.Close()
	os.Stdout = old
	var b bytes.Buffer
	io.Copy(&b, r)
	return b.String()
}

var escape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Every verb fits eighty columns, writes only background paints and
// resets, and under NO_COLOR writes no escape at all.
func TestVerbsAssumeNothingAboutTheTerminal(t *testing.T) {
	verbs := map[string]func() error{
		"known":       func() error { return known("") },
		"known-paint": func() error { return known("--paint") },
		"roundtrip":   roundtrip,
		"show":        func() error { return show("#ff6ec7") },
		"ramp": func() error {
			return ramp("#160d2b", "#ffa2ff",
				24)
		},
	}
	allowed := regexp.MustCompile(`^\x1b\[(48;2;\d+;\d+;\d+|0)m$`)
	for name, f := range verbs {
		out := capture(t, f)
		for n, line := range strings.Split(escape.ReplaceAllString(out,
			""), "\n") {
			if len(line) > 80 {
				t.Errorf("%s: line %d is %d columns", name, n+1,
					len(line))
			}
		}
		for _, e := range escape.FindAllString(out, -1) {
			if !allowed.MatchString(e) {
				t.Errorf("%s writes %q", name, e)
			}
		}
	}
	t.Setenv("NO_COLOR", "1")
	for name, f := range verbs {
		if escape.MatchString(capture(t, f)) {
			t.Errorf("%s writes escapes under NO_COLOR", name)
		}
	}
}

// The plain css from known is a :root block a browser would read.
func TestKnownCSSIsCSS(t *testing.T) {
	out := capture(t, func() error { return known("--css") })
	if !strings.HasPrefix(out, ":root {\n") || !strings.HasSuffix(out,
		"}\n") {
		t.Errorf("known --css is not one :root block:\n%s", out)
	}
	if escape.MatchString(out) {
		t.Errorf("known --css carries escapes; it is for piping into " +
			"a file")
	}
}

// Every wavelength across a profile falls in its own cell of the bar, so a
// round trip that lands elsewhere is the colour's loss, not the bar's.
func TestAcrossLandsInItsOwnCell(t *testing.T) {
	for _, profile := range radiation.Profiles {
		for i, nm := range across(profile) {
			if got := cellOf(profile, nm); got != i {
				t.Errorf("%s: %g nm is cell %d, want %d",
					profile.Name, nm, got, i)
			}
		}
		if cellOf(profile, profile.Source.Hi*2) != -1 {
			t.Errorf("%s: past the source is not -1", profile.Name)
		}
	}
}
