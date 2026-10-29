package colourway_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// grey is the one colour every role of the test source has.
const grey = "#808080"

// dusk is a source of one grey, under a name.
func dusk(name string) colourway.Source {
	roles := css.New()
	roles.Set("ground", srgb.MustHex(grey).Swatch())
	roles.Set("ink", srgb.MustHex(grey).Swatch())
	for _, each := range ghostty.Ansi {
		roles.Set(each, srgb.MustHex(grey).Swatch())
	}
	return colourway.Source{Name: name, About: "a test", Roles: roles,
		Origin: "sources/" + name + ".css"}
}

// sheet is a source file's text: the opening words and one role.
const sheet = "/*\n * a test\n */\n:root {\n  --ink: #808080;\n}\n"

// saved writes a source file with the base name, in a folder of its own,
// and returns the path.
func saved(t *testing.T, base string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), base)
	if err := os.WriteFile(path, []byte(sheet), private); err != nil {
		t.Fatal(err)
	}
	return path
}

// private is the permission a test file is written with. It is defined here.
const private = os.FileMode(0o600)

// nothing is how many entries a refused render leaves in its folder. It is
// defined here.
const nothing = 0

// The anti-test: a source whose file name would end up in the comments and
// quotes of the rendered files is refused when it is read, not when it is
// loaded in an editor.
func TestReadRefusesAFileNameItCannotCarry(t *testing.T) {
	for _, base := range []string{"a\nb.css", "a\rb.css", "a\x1bb.css", ".css"} {
		if _, err := colourway.Read(saved(t, base)); err == nil {
			t.Errorf("a source named %q was read", base)
		}
	}
}

// An ordinary file name with a quote and a space is a name: the writers
// quote it.
func TestReadTakesAQuotedFileName(t *testing.T) {
	source, err := colourway.Read(saved(t, "it's dusk.css"))
	if err != nil {
		t.Fatal(err)
	}
	if source.Name != "it's dusk" {
		t.Errorf("the name is %q", source.Name)
	}
}

// The anti-test: a source named with a line break, a slash or a dot-dot is
// refused for every program, its twin too, and nothing is written anywhere
// under the folder.
func TestRenderRefusesAName(t *testing.T) {
	for _, name := range []string{"a\nb", "../escape", "a/b", "..", ""} {
		dir := t.TempDir()
		for _, each := range colourway.Render(dusk(name), dir) {
			if each.Err == nil {
				t.Errorf("%s was written for the name %q", each.Program, name)
			}
		}
		left, err := os.ReadDir(dir)
		if err != nil || len(left) != nothing {
			t.Errorf("the name %q left %d entries in the folder: %v",
				name, len(left), err)
		}
	}
}

// A name with quotes goes through to the vim file quoted.
func TestRenderQuotesAName(t *testing.T) {
	dir := t.TempDir()
	name := "it's \"dusk\""
	for _, each := range colourway.Render(dusk(name), dir) {
		if each.Err != nil {
			t.Fatalf("%s: %v", each.Program, each.Err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "vim", "colors", name+".vim"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "let g:colors_name = 'it''s \"dusk\"'\n") {
		t.Error("the vim file does not hold the name quoted")
	}
}

// A source with a colour that is not a complete colour is refused when it
// is read, naming the line, and not rendered without the role.
func TestReadRefusesAColourItCannotRead(t *testing.T) {
	text := "/* a test */\n:root {\n  --ink: #808080;\n  --ground: #12;\n}\n"
	path := filepath.Join(t.TempDir(), "broken.css")
	if err := os.WriteFile(path, []byte(text), private); err != nil {
		t.Fatal(err)
	}
	_, err := colourway.Read(path)
	if err == nil || !strings.Contains(err.Error(), "line 4: --ground: #12") {
		t.Errorf("a source with a short colour gave %v", err)
	}
}

// A comment is a comment: words that close it do not let the rest of the
// words become roles when the source is written and read back.
func TestWordsCannotCloseTheirComment(t *testing.T) {
	words := "a test */ :root { --injected: #ff0000; } /* more"
	path := filepath.Join(t.TempDir(), "words.css")
	sheet := colourway.Commented(words) + ":root { --ink: #808080; }\n"
	if err := os.WriteFile(path, []byte(sheet), private); err != nil {
		t.Fatal(err)
	}
	back, err := colourway.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, has := back.Roles.Get("injected"); has {
		t.Error("a role came out of the opening words")
	}
	if !strings.Contains(back.About, "a test") {
		t.Errorf("the words came back as %q", back.About)
	}
}

// A render says which files it left out because the source cannot fill the
// program, and which it failed to write. A bare source is all skips; a
// source into a place that cannot be written is all failures.
func TestRenderSeparatesSkipsFromFailures(t *testing.T) {
	var bare colourway.Source
	bare.Name, bare.Roles = "bare", css.New()
	for _, each := range colourway.Render(bare, t.TempDir()) {
		if each.Err == nil || !each.Skipped() {
			t.Errorf("%s for %s: skipped %v, error %v", each.Program, each.Name,
				each.Skipped(), each.Err)
		}
	}
	blocked := t.TempDir()
	for _, program := range []string{"ghostty", "nvim", "vim", "glamour"} {
		if err := os.WriteFile(filepath.Join(blocked, program), nil, private); err != nil {
			t.Fatal(err)
		}
	}
	for _, each := range colourway.Render(dusk("dusk"), blocked) {
		if each.Err == nil || each.Skipped() {
			t.Errorf("%s for %s: skipped %v, error %v", each.Program, each.Name,
				each.Skipped(), each.Err)
		}
	}
	for _, each := range colourway.Render(dusk("a/b"), t.TempDir()) {
		if each.Err == nil || each.Skipped() {
			t.Errorf("a bad name was a skip, not a failure: %v", each.Err)
		}
	}
}

// A render that wrote is not skipped.
func TestAFileWrittenIsNotSkipped(t *testing.T) {
	for _, each := range colourway.Render(dusk("dusk"), t.TempDir()) {
		if each.Err != nil || each.Skipped() {
			t.Errorf("%s: %v", each.Program, each.Err)
		}
	}
}

// Rendering reads the source and changes none of it, however many files and
// twins it makes from it.
func TestRenderLeavesTheSourceAlone(t *testing.T) {
	source := dusk("dusk")
	before := source.Roles.String()
	colourway.Render(source, t.TempDir())
	if after := source.Roles.String(); after != before {
		t.Errorf("rendering changed the source:\n%s\nwas\n%s", after, before)
	}
}

// colours is the set of hex colours in a file.
func colours(t *testing.T, path string) map[string]bool {
	t.Helper()
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	set := map[string]bool{}
	lower := strings.ToLower(string(text))
	for _, each := range hexes.FindAllString(lower, all) {
		set[each] = true
	}
	return set
}

// The editor's two files come from one scheme: vim's file holds no colour
// that neovim's does not, for the source and for its twin.
func TestTheTwoEditorFilesHoldTheSameColours(t *testing.T) {
	dir := t.TempDir()
	colourway.Render(dusk("dusk"), dir)
	for _, name := range []string{"dusk", "dusk-inverted"} {
		lua := colours(t, filepath.Join(dir, "nvim", "colors", name+".lua"))
		vim := colours(t, filepath.Join(dir, "vim", "colors", name+".vim"))
		if len(vim) == nothing || len(lua) == nothing {
			t.Fatalf("%s: a file held no colour", name)
		}
		for each := range vim {
			if !lua[each] {
				t.Errorf("%s: vim holds %s and neovim does not", name, each)
			}
		}
	}
}

// A source the editor cannot be made from is skipped for both of its files,
// with one reason, and its terminal file is judged on its own.
func TestTheTwoEditorFilesAreSkippedTogether(t *testing.T) {
	var bare colourway.Source
	bare.Name, bare.Roles = "bare", css.New()
	reasons := map[string]map[string]string{}
	for _, each := range colourway.Render(bare, t.TempDir()) {
		if reasons[each.Name] == nil {
			reasons[each.Name] = map[string]string{}
		}
		if each.Err == nil {
			t.Fatalf("%s for %s was written", each.Program, each.Name)
		}
		reasons[each.Name][each.Program] = each.Err.Error()
	}
	for name, byProgram := range reasons {
		if byProgram["nvim"] == "" || byProgram["nvim"] != byProgram["vim"] {
			t.Errorf("%s: nvim says %q, vim says %q", name, byProgram["nvim"],
				byProgram["vim"])
		}
	}
}
