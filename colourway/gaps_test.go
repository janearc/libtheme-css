package colourway_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// numbered is the number claude-text is drawn in, written as a role. It is
// defined by the claude dialect's Slots.
const numbered = "palette-252"

// closer is the mark that ends a comment. It is defined by css.
const closer = "*/"

// once is how many times a comment closes. It is defined by this test.
const once = 1

// pale is a colour for the roles these tests add. It is defined by this test.
const pale = "#445566"

// Read of a missing file gives the platform's error.
func TestReadRefusesAMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nowhere.css")
	_, err := colourway.Read(missing)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a missing source gave %v", err)
	}
}

// readAbout is the opening words of a source file with this text.
func readAbout(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "opening.css")
	if err := os.WriteFile(path, []byte(text), private); err != nil {
		t.Fatal(err)
	}
	source, err := colourway.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	return source.About
}

// A source that does not open with a complete comment has no opening words.
// The anti-test is a comment that does close: it gives its words.
func TestASourceNeedsACompleteCommentToOpen(t *testing.T) {
	for _, text := range []string{"", ":root {}", "/* never closed", "x /* a */"} {
		if got := readAbout(t, text); got != "" {
			t.Errorf("%q opened with %q", text, got)
		}
	}
	if got := readAbout(t, "/* a */ :root {}"); got != "a" {
		t.Errorf("a closed comment gave %q", got)
	}
}

// No words, no comment. The anti-test is that some words are one.
func TestCommentedIsNothingForNoWords(t *testing.T) {
	if got := colourway.Commented(""); got != "" {
		t.Errorf("no words were commented as %q", got)
	}
	if got := colourway.Commented("a"); !strings.HasPrefix(got, "/*") {
		t.Errorf("words were commented as %q", got)
	}
}

// The property: a comment closes once, at its end, whatever words are in it.
func TestCommentedClosesOnce(t *testing.T) {
	rapid.Check(t, commentCloses)
}

// commentCloses is one draw of the property above.
func commentCloses(draw *rapid.T) {
	about := rapid.StringMatching(`[a-z*/ \n]+`).Draw(draw, "about")
	got := colourway.Commented(about)
	if strings.Count(got, closer) != once ||
		!strings.HasSuffix(got, closer+"\n") {
		draw.Fatalf("%q was commented as %q", about, got)
	}
}

// A colour set by name and by number is set twice. The terminal theme is
// skipped and says why. The anti-test is the same source set once: written.
func TestTerminalThemeRefusesAColourSetTwice(t *testing.T) {
	source := dusk("twice")
	source.Roles.Set("claude-text", srgb.MustHex(pale).Swatch())
	if clean, _ := renderOf(t, source, "ghostty"); clean.Err != nil {
		t.Errorf("set once it failed: %v", clean.Err)
	}
	source.Roles.Set(numbered, srgb.MustHex(pale).Swatch())
	got, _ := renderOf(t, source, "ghostty")
	if got.Program != "ghostty" || !got.Skipped() ||
		!strings.Contains(got.Err.Error(), "set twice") {
		t.Errorf("ghostty: skipped %v, %v", got.Skipped(), got.Err)
	}
}

// renderOf is what Render did for one program, for a source, and the folder
// it rendered into.
func renderOf(t *testing.T, source colourway.Source,
	program string) (colourway.Written, string) {
	t.Helper()
	dir := t.TempDir()
	for _, each := range colourway.Render(source, dir) {
		if each.Program == program {
			return each, dir
		}
	}
	t.Fatalf("render said nothing of %s", program)
	return colourway.Written{}, dir
}

// A markdown role gives glamour a style to write. The anti-test is the same
// colourway without one: glamour is skipped.
func TestAMarkdownRoleGivesAStyle(t *testing.T) {
	roles := css.New()
	roles.Set("ground", srgb.MustHex(grey).Swatch())
	roles.Set("ink", srgb.MustHex(grey).Swatch())
	bare := colourway.Source{Name: "bare", Roles: roles, Origin: "sources/bare.css"}
	if left, _ := renderOf(t, bare, "glamour"); !left.Skipped() {
		t.Errorf("glamour without a markdown role: %v", left.Err)
	}
	roles.Set("md-body", srgb.MustHex(pale).Swatch())
	made, dir := renderOf(t, bare, "glamour")
	if made.Err != nil {
		t.Errorf("glamour with a markdown role: %v", made.Err)
	}
	if _, err := os.Stat(filepath.Join(dir, "glamour", "bare.json")); err != nil {
		t.Error(err)
	}
}

// A markdown role with no ink and no md-body beside it makes no style. The
// file is skipped, not failed. The anti-test is that nothing is written.
func TestAMarkdownRoleWithoutInkIsSkipped(t *testing.T) {
	roles := css.New()
	roles.Set("md-code", srgb.MustHex(pale).Swatch())
	source := colourway.Source{Name: "odd", Roles: roles, Origin: "sources/odd.css"}
	left, dir := renderOf(t, source, "glamour")
	if !left.Skipped() || left.Path != "" {
		t.Errorf("path %q, error %v", left.Path, left.Err)
	}
	if _, err := os.Stat(filepath.Join(dir, "glamour")); err == nil {
		t.Error("a glamour folder was made for a style that was not written")
	}
}
