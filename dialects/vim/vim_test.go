package vim

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// scheme is a scheme with every colour neovim's scheme names, each its
// own grey, and a sixteen, so a group's colour can be told by its hex.
func scheme(t *testing.T) nvim.Scheme {
	t.Helper()
	made := nvim.Scheme{Name: "test", Colours: map[string]swatch.Swatch{}}
	for index, role := range nvim.Roles {
		level := uint8(16 + index*12)
		made.Colours[role.Name] = srgb.RGB8(level, level, level).Swatch()
	}
	for number := range 16 {
		made.Terminal = append(made.Terminal, srgb.RGB8(uint8(number*16), 0, 0).Swatch())
	}
	return made
}

// TestWriteSaysTheSchemeInVimscript finds the scheme named, the ground
// and ink on Normal, comments italic, no tree-sitter groups, and the
// sixteen for vim's terminal.
func TestWriteSaysTheSchemeInVimscript(t *testing.T) {
	var out strings.Builder
	if err := Write(&out, scheme(t), "test -- a scheme for the test"); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"let g:colors_name = 'test'",
		"hi Normal guifg=#1c1c1c guibg=#101010 gui=NONE cterm=NONE",
		"hi Comment guifg=#282828 gui=italic cterm=italic",
		"let g:terminal_ansi_colors = [",
		"'#f00000']",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the scheme does not say %q", want)
		}
	}
	if strings.Contains(text, "hi @") {
		t.Error("the scheme names a tree-sitter group, which vim cannot hold")
	}
}

// TestGroupsAreRead finds neovim's groups read from its own file, with
// their colours and styles.
func TestGroupsAreRead(t *testing.T) {
	for _, group := range nvim.Groups() {
		if group.Name == "CursorLineNr" {
			if group.Fg != "heading" || len(group.Styles) != 1 || group.Styles[0] != "bold" {
				t.Errorf("CursorLineNr read as %+v", group)
			}
			return
		}
	}
	t.Error("no CursorLineNr among the groups")
}

// TestEveryGroupIsRead finds as many groups as groups.lua has lines that
// set one, each colour a name the scheme has.
func TestEveryGroupIsRead(t *testing.T) {
	names := map[string]bool{"": true}
	for _, role := range nvim.Roles {
		names[role.Name] = true
	}
	groups := nvim.Groups()
	if len(groups) != 82 {
		t.Errorf("%d groups read; groups.lua sets 82", len(groups))
	}
	for _, group := range groups {
		if !names[group.Fg] || !names[group.Bg] {
			t.Errorf("%s has a colour the scheme does not name: %q %q",
				group.Name, group.Fg, group.Bg)
		}
	}
}

// TestLightSchemeSaysLight sets vim's background light for a scheme
// whose ground is lighter than its ink.
func TestLightSchemeSaysLight(t *testing.T) {
	light := scheme(t)
	light.Light = true
	var out strings.Builder
	if err := Write(&out, light, "light"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "set background=light") {
		t.Error("a light scheme did not say so")
	}
}
