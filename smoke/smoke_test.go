package smoke_test

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/dialects/vim"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// vimBreaker is a name that, raw in vimscript, closes the quote, ends the
// command and sets a variable. Setting a variable is harmless, and it is
// how these tests tell that the name escaped.
const vimBreaker = "x' | let g:injected = 1 | let g:tail = 'y"

// luaBreaker is a name that closes a lua string and opens the next.
const luaBreaker = `x"; injected = 1; y = "z`

// grey is the one colour every role of the test scheme has.
const grey = "#808080"

// dusk is a scheme of one grey, under a name.
func dusk(t *testing.T, name string) nvim.Scheme {
	t.Helper()
	roles := css.New()
	for _, entry := range nvim.Roles {
		roles.Set(entry.Role, srgb.MustHex(grey).Swatch())
	}
	for _, each := range ghostty.Ansi {
		roles.Set(each, srgb.MustHex(grey).Swatch())
	}
	scheme, err := nvim.Of(name, roles)
	if err != nil {
		t.Fatal(err)
	}
	return scheme
}

// A real vim sources the file written for a name that tries to run a
// command, and a header that tries to, and runs neither. The name it ends
// up with is the name that was given.
func TestVimLoadsAHostileSchemeAsAScheme(t *testing.T) {
	var out strings.Builder
	err := vim.Write(&out, dusk(t, vimBreaker), "words\rlet g:injected = 1")
	if err != nil {
		t.Fatal(err)
	}
	loaded := inVim(t, out.String())
	if loaded.injected {
		t.Error("vim ran a command that came from the name or the header")
	}
	if loaded.name != vimBreaker {
		t.Errorf("vim named the scheme %q, not %q", loaded.name, vimBreaker)
	}
}

// The same for neovim and a name that quotes lua.
func TestNeovimLoadsAHostileSchemeAsAScheme(t *testing.T) {
	var out strings.Builder
	err := dusk(t, luaBreaker).Write(&out, "words\rinjected = 1")
	if err != nil {
		t.Fatal(err)
	}
	loaded := inNvim(t, out.String())
	if loaded.injected {
		t.Error("neovim ran a statement that came from the name or the header")
	}
	if loaded.name != luaBreaker {
		t.Errorf("neovim named the scheme %q, not %q", loaded.name, luaBreaker)
	}
}

// spaced are names go's %q would write with \u escapes. They hold a no-break
// space, an em space and letters beyond ascii. LuaJIT does not read \uXXXX,
// so neovim has to be given them as they are.
var spaced = []string{"no\u00a0break", "em\u2003space", "caf\u00e9 \u5915\u66ae\u308c"}

// Neovim loads a scheme whose name has characters go would escape, and
// holds the name it was given.
func TestNeovimKeepsANameWithWideSpaces(t *testing.T) {
	for _, name := range spaced {
		var out strings.Builder
		if err := dusk(t, name).Write(&out, "header"); err != nil {
			t.Fatal(err)
		}
		loaded := inNvim(t, out.String())
		if loaded.injected || loaded.name != name {
			t.Errorf("neovim loaded %q as %+v", name, loaded)
		}
	}
}

// The smoke tests can pass: an ordinary scheme loads and names itself in
// both editors.
func TestBothEditorsLoadAnOrdinaryScheme(t *testing.T) {
	var vimText strings.Builder
	if err := vim.Write(&vimText, dusk(t, "dusk"), "header"); err != nil {
		t.Fatal(err)
	}
	var luaText strings.Builder
	if err := dusk(t, "dusk").Write(&luaText, "header"); err != nil {
		t.Fatal(err)
	}
	if loaded := inVim(t, vimText.String()); loaded.injected || loaded.name != "dusk" {
		t.Errorf("vim loaded an ordinary scheme as %+v", loaded)
	}
	if loaded := inNvim(t, luaText.String()); loaded.injected || loaded.name != "dusk" {
		t.Errorf("neovim loaded an ordinary scheme as %+v", loaded)
	}
}

// The anti-test: a file that does set the variable is seen setting it, so
// a pass above means the variable was not set and not that it went
// unseen.
func TestTheProbeSeesAnInjection(t *testing.T) {
	if !inVim(t, "let g:injected = 1\n").injected {
		t.Error("vim set the variable and the probe did not see it")
	}
	if !inNvim(t, "injected = 1\n").injected {
		t.Error("neovim set the variable and the probe did not see it")
	}
}
