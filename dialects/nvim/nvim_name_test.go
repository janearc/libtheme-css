package nvim_test

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/internal/plain"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// quoteBreaker is a name that closes a lua string and opens the next one.
// Set to nothing, it shows the name survives.
const quoteBreaker = `x"; injected = 1; y = "z`

// grey is the one colour every role of the test scheme has.
const grey = "#808080"

// roles is a colourway of one grey.
func roles() *css.Sheet {
	sheet := css.New()
	for _, entry := range nvim.Roles {
		sheet.Set(entry.Role, srgb.MustHex(grey).Swatch())
	}
	for _, name := range ghostty.Ansi {
		sheet.Set(name, srgb.MustHex(grey).Swatch())
	}
	return sheet
}

// dusk is a scheme of one grey.
func dusk(t interface{ Fatal(...any) }, name string) nvim.Scheme {
	scheme, err := nvim.Of(name, roles())
	if err != nil {
		t.Fatal(err)
	}
	return scheme
}

// written is a scheme and a header as lua.
func written(scheme nvim.Scheme, header string) (string, error) {
	var out strings.Builder
	err := scheme.Write(&out, header)
	return out.String(), err
}

// The anti-test: a name with a line break or another control character, or
// that is no name, is not made into a scheme or written, and nothing is
// written before the refusal. Lua ends a comment at a carriage return as
// well as at a newline.
func TestANameThatIsNoNameIsRefused(t *testing.T) {
	for _, name := range []string{"x\ninjected = 1", "a\rb", "a\x00b", "a\x1bb",
		"a\u2028b", "", "..", "a/b", "a\u200bb", "a\u00adb"} {
		scheme := dusk(t, "dusk")
		scheme.Name = name
		if text, err := written(scheme, "header"); err == nil || text != "" {
			t.Errorf("the name %q was written: %v", name, err)
		}
		if _, err := nvim.Of(name, roles()); err == nil {
			t.Errorf("Of made a scheme named %q", name)
		}
	}
}

func TestAHeaderCannotLeaveItsComment(t *testing.T) {
	header := "one\rinjected = 1\r\ntwo\vinjected = 1\u2028injected = 1"
	text, err := written(dusk(t, "dusk"), header)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "injected") && !strings.HasPrefix(line, "--") {
			t.Errorf("header text became code: %q", line)
		}
	}
	if strings.ContainsAny(text, "\r\v\f\u2028") {
		t.Error("a line break other than a newline reached the file")
	}
}

// A name with quotes is written as lua holds them and read back as the name
// it was.
func TestAQuotedNameReadsBack(t *testing.T) {
	text, err := written(dusk(t, quoteBreaker), "header")
	if err != nil {
		t.Fatal(err)
	}
	again, err := nvim.Read(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if again.Name != quoteBreaker {
		t.Errorf("the name read back as %q, not %q", again.Name, quoteBreaker)
	}
}

// wide are names that go's %q writes with \u escapes. LuaJIT does not read
// them. They hold a no-break space, an em space and letters beyond ascii.
var wide = []string{"no\u00a0break", "em\u2003space", "caf\u00e9 \u5915\u66ae\u308c"}

// Names are written as lua takes them, each character as it is. The
// anti-test is a name with a quote, the one character that gets a backslash.
func TestAWideNameIsWrittenAsItIs(t *testing.T) {
	for _, name := range append(wide, quoteBreaker) {
		text, err := written(dusk(t, name), "header")
		if err != nil {
			t.Fatal(err)
		}
		want := `vim.g.colors_name = "` + strings.ReplaceAll(name, `"`, `\"`) + `"`
		if !strings.Contains(text, want) {
			t.Errorf("no %q in the file for the name %q", want, name)
		}
		if strings.Contains(text, `\u`) {
			t.Errorf("the file for the name %q holds a \\u escape", name)
		}
	}
}

// graphicName is a name the library accepts. It is text of graphic
// characters, and not a place or a path.
func graphicName() *rapid.Generator[string] {
	return rapid.String().Filter(isAName)
}

// isAName is whether plain.Name takes a name.
func isAName(name string) bool { return plain.Name(name) == nil }

// readsBack is the property: a name written is read back as it was.
func readsBack(t *rapid.T) {
	name := graphicName().Draw(t, "name")
	text, err := written(dusk(t, name), "header")
	if err != nil {
		t.Fatal(err)
	}
	again, err := nvim.Read(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	if again.Name != name {
		t.Fatalf("the name read back as %q, not %q", again.Name, name)
	}
}

// Whatever name the library accepts is written and read back as it was.
func TestAnyNameReadsBack(t *testing.T) {
	rapid.Check(t, readsBack)
}

// tabled is a colour table and a name line for the reading tests.
const tabled = "\nlocal c = {\n  bg = \"#000000\",\n}\n"

// Names are read as lua spells them. Escaped quotes are quotes. Double dashes
// inside the quotes are part of the name and not a comment. Escapes go does
// not know leave the name as written.
//
// That is go's reading and not lua's: lua reads \z as a skip over white
// space. The reader is for files this library wrote. Their names hold no
// backslash, so \" is the only escape in them.
func TestAHandWrittenNameIsRead(t *testing.T) {
	cases := map[string]string{
		`vim.g.colors_name = "a\"b"`:         `a"b`,
		`vim.g.colors_name = "a--b" -- note`: "a--b",
		`vim.g.colors_name = "a\zb"`:         `a\zb`,
	}
	for line, want := range cases {
		scheme, err := nvim.Read(strings.NewReader(line + tabled))
		if err != nil {
			t.Fatal(err)
		}
		if scheme.Name != want {
			t.Errorf("%s read as %q, want %q", line, scheme.Name, want)
		}
	}
}

// The counts of colours in a terminal table the tests below read. Each is
// defined by the terminal, which has sixteen: eight and their bright eight.
const (
	// sixteen is the colours a terminal table has. It is defined by the
	// terminal.
	sixteen = 16
	// tooMany is a table with one colour over. It is defined by sixteen.
	tooMany = 17
	// tooFew is a table with one colour short. It is defined by sixteen.
	tooFew = 15
	// entries is how many entries the stray-line file's colour table
	// holds. It is defined by that file: bg and ink.
	entries = 2
)

// termOf is a terminal table of some black colours, after the colour table.
func termOf(count int) string {
	return tabled + "local term = {\n" + strings.Repeat(`"#000000", `, count) +
		"\n}\n"
}

// readTerm reads a terminal table of some black colours.
func readTerm(count int) (nvim.Scheme, error) {
	text := termOf(count)
	return nvim.Read(strings.NewReader(text))
}

// The terminal table has sixteen colours or none. More, or fewer, is
// refused, and the message says how many there were.
func TestTheSixteenAreSixteenOrNone(t *testing.T) {
	for count, want := range map[int]string{tooMany: "more than sixteen",
		tooFew: "15 of the sixteen"} {
		_, err := readTerm(count)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%d colours: %v, want %q", count, err, want)
		}
	}
	if scheme, err := readTerm(sixteen); err != nil ||
		len(scheme.Terminal) != sixteen {
		t.Errorf("sixteen colours: %v, %d read", err, len(scheme.Terminal))
	}
}

// A file that cannot be read says so.
func TestAnUnreadableFileIsRefused(t *testing.T) {
	unreadable := iotest.ErrReader(errors.New("no file"))
	if _, err := nvim.Read(unreadable); err == nil {
		t.Error("an unreadable file was read")
	}
}

// A line in the colour table that is no entry is passed over, and the
// entries around it are read.
func TestAStrayLineInTheColoursIsPassedOver(t *testing.T) {
	stray := "local c = {\n  bg = \"#000000\",\n  not an entry\n" +
		"  ink = \"#ffffff\",\n}\n"
	scheme, err := nvim.Read(strings.NewReader(stray))
	if err != nil || len(scheme.Colours) != entries {
		t.Errorf("colours read: %v, %v", scheme.Colours, err)
	}
}

// A name that sits in a comment is not the scheme's.
func TestACommentedNameIsNotTheName(t *testing.T) {
	scheme, err := nvim.Read(strings.NewReader("-- vim.g.colors_name = \"no\"" + tabled))
	if err != nil {
		t.Fatal(err)
	}
	if scheme.Name != "" {
		t.Errorf("the name is %q", scheme.Name)
	}
}
