package nvim_test

import (
	"os"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// aGroupStarts is how a group begins in groups.lua. It is defined by that
// file: every group is one call of g with window 0.
const aGroupStarts = "g(0,"

// byName is the groups by their names, for a test to look in.
func byName(groups []nvim.Group) map[string]nvim.Group {
	found := map[string]nvim.Group{}
	for _, group := range groups {
		found[group.Name] = group
	}
	return found
}

// Groups are the groups groups.lua sets: one for each of its group lines,
// each with the names of its colours and its styles.
func TestGroupsAreWhatGroupsLuaSets(t *testing.T) {
	raw, err := os.ReadFile("groups.lua")
	if err != nil {
		t.Fatal(err)
	}
	want := nothing
	for line := range strings.SplitSeq(string(raw), "\n") {
		if strings.HasPrefix(line, aGroupStarts) {
			want++
		}
	}
	groups := nvim.Groups()
	found := byName(groups)
	if len(groups) != want || len(found) != want {
		t.Errorf("%d groups, %d of them named differently, %d lines", len(groups),
			len(found), want)
	}
	if normal := found["Normal"]; normal.Fg != "ink" || normal.Bg != "bg" {
		t.Errorf("Normal is %+v", normal)
	}
	if comment := found["Comment"]; comment.Fg != "dim" ||
		strings.Join(comment.Styles, ",") != "italic" {
		t.Errorf("Comment is %+v", comment)
	}
	if title := found["Title"]; strings.Join(title.Styles, ",") != "bold" {
		t.Errorf("Title is %+v", title)
	}
}

// The anti-test: no group names a colour the scheme's table lacks, since a
// name the table does not have would be nil in the file and the group would
// fall back to someone else's colour.
func TestNoGroupNamesAColourTheTableLacks(t *testing.T) {
	table := map[string]bool{}
	for _, entry := range nvim.Roles {
		table[entry.Name] = true
	}
	for _, group := range nvim.Groups() {
		for _, name := range []string{group.Fg, group.Bg} {
			if name != "" && !table[name] {
				t.Errorf("%s names %q, which the table lacks", group.Name, name)
			}
		}
	}
}

// The anti-test: what is not a group line is not a group. The comments,
// the preamble and the blank lines of the file make no group.
func TestOnlyGroupLinesMakeGroups(t *testing.T) {
	for _, group := range nvim.Groups() {
		if group.Name == "" || strings.ContainsAny(group.Name, " -{}") {
			t.Errorf("a group called %q", group.Name)
		}
	}
}

// The anti-test: a colourway without one of the sixteen is refused, by
// name, saying the terminal needs it, and no scheme comes back.
func TestOfRefusesAMissingTerminalColour(t *testing.T) {
	for _, name := range ghostty.Ansi {
		roles := roles()
		without := css.New()
		for _, rule := range roles.Rules() {
			if rule.Name != name {
				without.Set(rule.Name, rule.Swatch)
			}
		}
		scheme, err := nvim.Of("dusk", without)
		if err == nil || !strings.Contains(err.Error(), name) ||
			!strings.Contains(err.Error(), "the terminal") {
			t.Errorf("without %s: %v", name, err)
		}
		if len(scheme.Colours) != nothing || len(scheme.Terminal) != nothing {
			t.Errorf("without %s: a scheme came back", name)
		}
	}
}

// A colourway may say a colour for the editor alone: nvim-ink is the scheme's
// ink, and the terminal's ink is untouched. Where it says nothing, the
// role's own colour is the scheme's.
func TestAnEditorRoleWinsForTheEditorOnly(t *testing.T) {
	roles := roles()
	own := srgb.MustHex("#336699").Swatch()
	roles.Set("nvim-ink", own)
	scheme, err := nvim.Of("dusk", roles)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ghostty.Hex(scheme.Colours["ink"]),
		ghostty.Hex(own); got != want {
		t.Errorf("the ink is %s, not %s", got, want)
	}
	if got := ghostty.Hex(scheme.Colours["dim"]); got != grey {
		t.Errorf("the dim is %s, not the role's %s", got, grey)
	}
	if _, found := scheme.Roles().Get("nvim-ink"); found {
		t.Error("the editor-only role leaked into the roles")
	}
}
