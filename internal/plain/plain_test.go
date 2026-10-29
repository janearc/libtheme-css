package plain_test

import (
	"strings"
	"testing"
	"unicode"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/internal/plain"
)

// ordinary are names with spaces, quotes, dots, dashes and non-ascii
// letters: legal file names, and each writer can quote them.
var ordinary = []string{"dusk", "it's dusk", `say "dusk"`, "a.b", "-dusk",
	"dusk-inverted", "crépuscule", "夕暮れ", ".hidden", "no\u00a0break",
	"em\u2003space"}

// refused are names that are no names. Some are empty, a place or a path.
// Some are invalid text. Some hold a control character or an invisible one,
// such as a soft hyphen or a zero width space.
var refused = []string{"", ".", "..", "a/b", `a\b`, "a\nb", "a\rb", "a\vb",
	"a\fb", "a\u0085b", "a\u2028b", "a\u2029b", "a\x00b", "a\x1bb",
	"a\x7fb", "a\xffb", "a\u0080b", "a\tb", "a\u00adb", "a\u200bb",
	"a\u2060b", "a\ufeffb", "a\u200eb"}

func TestNameAcceptsOrdinaryNames(t *testing.T) {
	for _, name := range ordinary {
		if err := plain.Name(name); err != nil {
			t.Errorf("%q was refused: %v", name, err)
		}
	}
}

// The anti-test: each name that is no name for a colourway is refused.
func TestNameRefuses(t *testing.T) {
	for _, name := range refused {
		if plain.Name(name) == nil {
			t.Errorf("%q was accepted", name)
		}
	}
}

func TestCommentBreaksAtEveryLineEnd(t *testing.T) {
	words := "a\nb\rc\r\nd\ve\ff\u0085g\u2028h\u2029i\tj"
	got := plain.Comment("#", words)
	want := "# a\n# b\n# c\n# d\n# e\n# f\n# g\n# h\n# i\tj\n"
	if got != want {
		t.Errorf("Comment(%q) = %q, want %q", words, got, want)
	}
}

func TestCommentDropsOtherControlsAndTrims(t *testing.T) {
	got := plain.Comment("#", "  a\x00b\x1bc\x7fd  ")
	if got != "# abcd\n" {
		t.Errorf("Comment gave %q", got)
	}
}

// Empty words are one bare mark, which is what a header of nothing was.
func TestCommentOfNothingIsTheBareMark(t *testing.T) {
	if got := plain.Comment("--", "  \n "); got != "--\n" {
		t.Errorf("Comment of nothing gave %q", got)
	}
}

// Ordinary multi-line words come out as they were, line for line, so a
// header already in use is written the same.
func TestCommentKeepsOrdinaryWords(t *testing.T) {
	got := plain.Comment("--", "one\n\ntwo words \nthree")
	want := "-- one\n--\n-- two words\n-- three\n"
	if got != want {
		t.Errorf("Comment gave %q, want %q", got, want)
	}
}

// The property: whatever the words, every line of the comment starts with
// the mark and holds no control character but the tab.
func TestCommentNeverLeavesTheMark(t *testing.T) {
	rapid.Check(t, keepsToTheMark)
}

// keepsToTheMark is one draw of the property above.
func keepsToTheMark(t *rapid.T) {
	words := rapid.String().Draw(t, "words")
	text := plain.Comment("#", words)
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if !strings.HasPrefix(line, "#") {
			t.Fatalf("Comment(%q) has the line %q", words, line)
		}
		if strings.ContainsFunc(line, stray) {
			t.Fatalf("Comment(%q) kept a stray character in %q", words, line)
		}
	}
}

// stray is a character a comment line must not hold.
func stray(each rune) bool {
	return each != '\t' && (unicode.IsControl(each) || each == '\u2028' || each == '\u2029')
}

// hidden is a character no name may hold. Unicode names its classes: C, Zl
// and Zp. It is worked out from them and not from IsGraphic.
func hidden(each rune) bool {
	return unicode.In(each, unicode.C, unicode.Zl, unicode.Zp)
}

// The property: a name is accepted exactly when it is text and graphic. It is
// not a place and not a path.
func TestNameAgreesWithItsRules(t *testing.T) {
	rapid.Check(t, nameAgrees)
}

// nameAgrees is one draw of the property above.
func nameAgrees(t *rapid.T) {
	name := rapid.String().Draw(t, "name")
	clean := name != "" && name != "." && name != ".." &&
		!strings.ContainsAny(name, `/\`) && !strings.ContainsFunc(name, hidden)
	if got := plain.Name(name) == nil; got != clean {
		t.Fatalf("Name(%q) accepted %v, want %v", name, got, clean)
	}
}
