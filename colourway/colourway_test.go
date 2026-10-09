package colourway

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// terminal is a source with what a terminal theme needs and nothing for
// an editor.
func terminal() Source {
	roles := css.New()
	roles.Set("ground", srgb.MustHex("#808a63").Swatch())
	roles.Set("ink", srgb.MustHex("#38241e").Swatch())
	for number, name := range ghostty.Ansi {
		roles.Set(name, srgb.RGB8(uint8(number*16), 40, 90).Swatch())
	}
	roles.Set("claude-text", srgb.MustHex("#c5e699").Swatch())
	return Source{Name: "test-dusk", About: "a test colourway\nin two lines",
		Roles: roles, Origin: "sources/test-dusk.css"}
}

// a source written and read back has its words and its roles.
func TestWriteRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-dusk.css")
	if err := terminal().Write(path); err != nil {
		t.Fatal(err)
	}
	back, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if back.Name != "test-dusk" || back.About != "a test colourway\nin two lines" {
		t.Errorf("read back as %q, %q", back.Name, back.About)
	}
	if strings.Join(back.Roles.Names(), " ") != strings.Join(terminal().Roles.Names(), " ") {
		t.Error("the roles came back different")
	}
}

// Render writes a terminal colourway for every program, twins too, since
// what the editor and markdown need is derived from the sixteen; the
// terminal theme has claude code's colour in its palette and says where
// it came from.
func TestRenderWritesEveryProgram(t *testing.T) {
	dir := t.TempDir()
	written := Render(terminal(), dir)
	wrote, skipped := 0, 0
	for _, each := range written {
		if each.Err != nil {
			skipped++
			continue
		}
		wrote++
	}
	if wrote != 8 || skipped != 0 {
		t.Errorf("wrote %d, skipped %d: %+v", wrote, skipped, written)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "ghostty", "test-dusk"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, want := range []string{"# a test colourway", "sources/test-dusk.css",
		"palette = 252=#c5e699"} {
		if !strings.Contains(text, want) {
			t.Errorf("no %q in the theme", want)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "ghostty", "test-dusk-inverted")); err != nil {
		t.Error("no inverted twin")
	}
	if _, err := os.Stat(filepath.Join(dir, "vim", "colors", "test-dusk.vim")); err != nil {
		t.Error("no scheme for vim")
	}
}

// a colourway with only a ground and an ink has no sixteen for the
// terminal or the editor, and nothing but its ink for markdown, so every
// program is skipped, markdown saying it had only the ink.
func TestRenderSkipsWhatItCannotFill(t *testing.T) {
	roles := css.New()
	roles.Set("ground", srgb.MustHex("#808a63").Swatch())
	roles.Set("ink", srgb.MustHex("#38241e").Swatch())
	for _, each := range Render(Source{Name: "bare", Roles: roles, Origin: "sources/bare.css"},
		t.TempDir()) {
		if each.Err == nil {
			t.Errorf("%s for %s was written", each.Program, each.Name)
		}
		if each.Program == "glamour" && !strings.Contains(each.Err.Error(), "but the ink") {
			t.Errorf("glamour for %s: %v", each.Name, each.Err)
		}
	}
}

// a source with no selection gets the derived one in its terminal theme,
// the same surface-1 the editor's scheme has, so what paratune shows on
// the shell page is what the terminal draws.
func TestRenderDerivesTheSelection(t *testing.T) {
	dir := t.TempDir()
	Render(terminal(), dir)
	selection, _ := Resolve(terminal().Roles).Of("surface-1")
	raw, err := os.ReadFile(filepath.Join(dir, "ghostty", "test-dusk"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "selection-background = " + ghostty.Hex(selection); !strings.Contains(string(raw), want) {
		t.Errorf("no %q in the theme", want)
	}
}

// the terminal theme carries claude code's colours the colourway does not
// set, derived, so claude code follows the colourway with none set.
func TestRenderDerivesClaude(t *testing.T) {
	dir := t.TempDir()
	Render(terminal(), dir)
	mode, _ := Resolve(terminal().Roles).Of("claude-mode")
	raw, err := os.ReadFile(filepath.Join(dir, "ghostty", "test-dusk"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "palette = 180=" + ghostty.Hex(mode); !strings.Contains(string(raw), want) {
		t.Errorf("no %q in the theme", want)
	}
}
