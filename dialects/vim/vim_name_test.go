package vim_test

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/dialects/vim"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// quoteBreaker is a colourway name that, raw in vimscript, closes the quote,
// ends the command and sets a variable. Setting a variable is harmless, and
// it is how the smoke test tells that the name escaped.
const quoteBreaker = "x' | let g:injected = 1 | let g:tail = 'y"

// grey is the one colour every role of the test scheme has.
const grey = "#808080"

// dusk is a scheme of one grey, under a name.
func dusk(t *testing.T, name string) nvim.Scheme {
	t.Helper()
	roles := css.New()
	for _, entry := range nvim.Roles {
		roles.Set(entry.Role, srgb.MustHex(grey).Swatch())
	}
	for _, name := range ghostty.Ansi {
		roles.Set(name, srgb.MustHex(grey).Swatch())
	}
	scheme, err := nvim.Of("dusk", roles)
	if err != nil {
		t.Fatal(err)
	}
	scheme.Name = name
	return scheme
}

// written is a scheme and a header as vimscript.
func written(t *testing.T, name, header string) (string, error) {
	t.Helper()
	var out strings.Builder
	err := vim.Write(&out, dusk(t, name), header)
	return out.String(), err
}

func TestANameIsQuotedForVimscript(t *testing.T) {
	text, err := written(t, quoteBreaker, "header")
	if err != nil {
		t.Fatal(err)
	}
	want := "let g:colors_name = 'x'' | let g:injected = 1 | let g:tail = ''y'\n"
	if !strings.Contains(text, want) {
		t.Errorf("the name is not quoted for vimscript; want %q", want)
	}
}

// The anti-test: a name with a line break or another control character, or
// that is no name, is refused and nothing is written.
func TestANameThatIsNoNameIsRefused(t *testing.T) {
	for _, name := range []string{"x\nlet g:injected = 1\n\"y", "a\rb", "a\x00b",
		"a\x1bb", "a\u2028b", "", "..", "a/b"} {
		if text, err := written(t, name, "header"); err == nil || text != "" {
			t.Errorf("the name %q was written: %v", name, err)
		}
	}
}

// A header's every kind of line end starts another comment line, so a
// header cannot put code in the file.
func TestAHeaderCannotLeaveItsComment(t *testing.T) {
	header := "one\rlet g:injected = 1\r\ntwo\vlet g:injected = 1\u2028let g:injected = 1"
	text, err := written(t, "dusk", header)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "g:injected") && !strings.HasPrefix(line, `"`) {
			t.Errorf("header text became code: %q", line)
		}
	}
	if strings.ContainsAny(text, "\r\v\f\u2028") {
		t.Error("a line break other than a newline reached the file")
	}
}

// The anti-test: vim is written from a complete scheme only, like neovim: a
// scheme with a colour missing or a short sixteen is an error and nothing is
// written, where it used to be written with groups quietly left out.
func TestAnIncompleteSchemeIsNotWritten(t *testing.T) {
	short := dusk(t, "dusk")
	short.Terminal = short.Terminal[:len(short.Terminal)/half]
	holy := dusk(t, "dusk")
	delete(holy.Colours, "panel")
	var empty nvim.Scheme
	empty.Name = "empty"
	for name, scheme := range map[string]nvim.Scheme{"short": short, "holy": holy, "empty": empty} {
		var out strings.Builder
		if err := vim.Write(&out, scheme, "header"); err == nil || out.Len() != nothing {
			t.Errorf("a %s scheme was written: %v", name, err)
		}
	}
}

// half is the middle of a table, and nothing is no bytes. Both are defined
// by counting.
const (
	// half cuts a table in two. It is defined by halving.
	half = 2
	// nothing is no bytes written. It is defined by counting.
	nothing = 0
)
