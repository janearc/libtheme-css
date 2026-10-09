package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// sheetOf is a sheet from names and hexes, in order.
func sheetOf(pairs ...string) *css.Sheet {
	sheet := css.New()
	for index := 0; index+1 < len(pairs); index += 2 {
		sheet.Set(pairs[index], srgb.MustHex(pairs[index+1]).Swatch())
	}
	return sheet
}

// a diff counts what is the same and says each role that changed, came
// or went, and a change too small to see says so.
func TestReport(t *testing.T) {
	before := sheetOf("ground", "#101010", "ink", "#d0d0d0", "panel", "#202020")
	after := sheetOf("ground", "#101010", "ink", "#d1d0d0", "heading", "#b893b4")
	var out bytes.Buffer
	if !report(&out, compared(before, after), false) {
		t.Error("a diff with changes said there were none")
	}
	text := out.String()
	for _, want := range []string{"1 the same, 1 changed, 1 added, 1 removed",
		"~ ink", "#d0d0d0 -> #d1d0d0", "under what an eye can tell",
		"+ heading            #b893b4", "- panel              #202020"} {
		if !strings.Contains(text, want) {
			t.Errorf("no %q in\n%s", want, text)
		}
	}
	out.Reset()
	if report(&out, compared(before, before), false) {
		t.Error("a colourway compared with itself differed")
	}
}

// a change an eye can see carries no note, and its hue goes the short way
// round the wheel.
func TestReportVisible(t *testing.T) {
	var out bytes.Buffer
	report(&out, compared(sheetOf("red", "#ff0000"), sheetOf("red", "#ff00ff")), false)
	if text := out.String(); strings.Contains(text, "under what") || !strings.Contains(text, "H -") {
		t.Errorf("a visible change reads %q", text)
	}
}

// git's /dev/null, for a file added or removed, is no roles.
func TestNothingFromDevNull(t *testing.T) {
	roles, err := rolesOf(os.DevNull)
	if err != nil || len(roles.Names()) != 0 {
		t.Errorf("/dev/null gave %v, %v", roles.Names(), err)
	}
}

// as git's driver it never fails, whatever git hands it: a rename's nine
// arguments, an unmerged path's one, a file it cannot read.
func TestGitDriverNeverFails(t *testing.T) {
	for _, arguments := range [][]string{
		{"a.css"},
		{"a.css", os.DevNull, "0", "0", os.DevNull, "0", "0"},
		{"a.css", os.DevNull, "0", "0", os.DevNull, "0", "0", "b.css", "similarity index 100%"},
		{"a.css", "/no/such/file.css", "0", "0", os.DevNull, "0", "0"},
	} {
		if differed, err := diff(append([]string{"--git"}, arguments...)); differed || err != nil {
			t.Errorf("%d arguments: %v, %v", len(arguments), differed, err)
		}
	}
}
