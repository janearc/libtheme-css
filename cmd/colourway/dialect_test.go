package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// private is the permission a test file is written with. It is defined
// here.
const private = os.FileMode(0o600)

// stored writes text to a file called base in a folder of its own and
// returns the path.
func stored(t *testing.T, base, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), base)
	if err := os.WriteFile(path, []byte(text), private); err != nil {
		t.Fatal(err)
	}
	return path
}

// inkOnly is the roles inkCSS has.
var inkOnly = []string{"ink"}

// inkCSS is a source with one role, in the css dialect.
const inkCSS = ":root {\n  --ink: #ffffff;\n}\n"

// The name picks the dialect, whatever the file on disk is called, and the
// extension is read in any case.
func TestTheNamePicksTheDialect(t *testing.T) {
	path := stored(t, "XXXXXX", inkCSS)
	for _, name := range []string{"a.css", "A.CSS", "dir/a.Css"} {
		roles, err := rolesOf(side{name: name, path: path})
		if err != nil || !slices.Equal(roles.Names(), inkOnly) {
			t.Errorf("%s read as %v, %v", name, roles.Names(), err)
		}
	}
	if dialectOf("a.lua").says != "a neovim scheme" ||
		dialectOf("a.json").says != "a markdown style" ||
		dialectOf("ghostty/a").says != "a terminal theme" ||
		dialectOf("a.dark").says != "a terminal theme" {
		t.Error("a name was given a dialect it does not say")
	}
}

// The anti-tests: text that gives no colours is an error that names the
// dialect it was read as; so is a file that cannot be read, or that its
// dialect refuses. Whitespace alone is an empty file, and is not.
func TestAFileOfTheWrongDialectIsAnError(t *testing.T) {
	_, err := rolesOf(side{name: "a.lua", path: stored(t, "x", inkCSS)})
	if err == nil {
		t.Error("css read as lua gave no error")
	}
	for name, want := range map[string]string{"a": "a terminal theme",
		"a.css": "a css source"} {
		text := "vim.g.colors_name = \"x\"\n"
		_, err = rolesOf(side{name: name, path: stored(t, "x", text)})
		if err == nil || !strings.Contains(err.Error(), "read as "+want) {
			t.Errorf("%s gave %v, not an error that says %s", name, err, want)
		}
	}
	_, err = rolesOf(side{name: "a.json", path: stored(t, "x", inkCSS)})
	if err == nil {
		t.Error("css read as json gave no error")
	}
	_, err = rolesOf(side{name: "a", path: stored(t, "x", "palette = 300=#000000\n")})
	if err == nil {
		t.Error("a terminal theme with a bad palette entry gave no error")
	}
	_, err = rolesOf(side{name: "a.css", path: "/no/such/file"})
	if err == nil {
		t.Error("a file that is not there gave no error")
	}
	roles, err := rolesOf(side{name: "a.css", path: stored(t, "x", " \n\t\n")})
	if err != nil || len(roles.Names()) != nothing {
		t.Errorf("a blank file gave %v, %v", roles.Names(), err)
	}
}

// Git's driver prints a line for each call shape and then the comparison.
func TestTheDriverNamesWhatItCompares(t *testing.T) {
	old, fresh := stored(t, "o", inkCSS), stored(t, "n", strings.ReplaceAll(inkCSS, "ffffff", "eeeeee"))
	cases := map[string][]string{
		"colourway diff x.css\n1 the same":            {"x.css", old, "a", "1", old, "b", "1"},
		"colourway diff x.css\n0 the same, 1 changed": {"x.css", old, "a", "1", fresh, "b", "1"},
		"colourway diff x.css -> y.css\n0 the same, 1 changed": {"x.css", old, "a", "1",
			fresh, "b", "1", "y.css", "similarity"},
		"colourway diff x.css: unmerged":       {"x.css"},
		"colourway diff: git gave 2 arguments": {"x.css", old},
		"colourway: x.css has text and no colours": {"x.css", old, "a", "1",
			stored(t, "z", "nothing here\n"), "b", "1"},
	}
	for want, arguments := range cases {
		var out bytes.Buffer
		gitDiff(&out, arguments, false)
		if !strings.Contains(out.String(), want) {
			t.Errorf("%v printed %q, not %q", arguments, out.String(), want)
		}
	}
}
