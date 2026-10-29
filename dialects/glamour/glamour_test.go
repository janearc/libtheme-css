package glamour

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// editor is the roles an editor's colourway already has, and no
// markdown ones.
func editor() *css.Sheet {
	roles := css.New()
	for name, hex := range map[string]string{"ink": "#38241e", "heading": "#3a1830",
		"link": "#141c3a", "code": "#1c2810", "string": "#3a2c08", "dim": "#2e2a1c",
		"panel": "#8f8856", "border": "#6e683e", "err": "#4a1810"} {
		roles.Set(name, srgb.MustHex(hex).Swatch())
	}
	return roles
}

// colours are every colour a style sets, by dotted key.
func colours(t *testing.T, style []byte) map[string]string {
	t.Helper()
	parsed := map[string]any{}
	if err := json.Unmarshal(style, &parsed); err != nil {
		t.Fatal(err)
	}
	found := map[string]string{}
	var walk func(node any, path string)
	walk = func(node any, path string) {
		switch value := node.(type) {
		case map[string]any:
			for key, child := range value {
				walk(child, strings.TrimPrefix(path+"."+key, "."))
			}
		case string:
			if strings.HasSuffix(path, "color") {
				found[path] = value
			}
		}
	}
	walk(parsed, "")
	return found
}

// every colour in base.json is a key some role sets, so nothing of the
// base's own colourway is left in another's style.
func TestEveryBaseColourHasARole(t *testing.T) {
	keyed := map[string]bool{}
	for _, entry := range Roles {
		for _, key := range entry.Keys {
			keyed[key] = true
		}
	}
	for key := range colours(t, base) {
		if !keyed[key] {
			t.Errorf("%s is in base.json and no role sets it", key)
		}
	}
}

// an editor's colourway makes a style that agrees with its vim:
// headings in heading, links in link, comments in dim, inline code on
// the panel, and no ground where there is none to fall back to.
func TestAnEditorsRolesMakeAStyle(t *testing.T) {
	style, err := Of(editor())
	if err != nil {
		t.Fatal(err)
	}
	got := colours(t, style)
	for key, want := range map[string]string{"heading.color": "#3a1830",
		"h3.color": "#141c3a", "document.color": "#38241e",
		chroma + "comment.color": "#2e2a1c", chroma + "keyword.color": "#3a1830",
		chroma + "literal_string.color": "#3a2c08", "code.background_color": "#8f8856",
		"emph.color": "#38241e", chroma + "error.color": "#4a1810"} {
		if got[key] != want {
			t.Errorf("%s is %q, want %s", key, got[key], want)
		}
	}
	if _, found := got["heading.background_color"]; found {
		t.Error("the heading has a ground nobody gave it")
	}
}

// a markdown role wins over the editor's role it would fall back to.
func TestAMarkdownRoleWins(t *testing.T) {
	roles := editor()
	roles.Set("md-heading", srgb.MustHex("#112233").Swatch())
	roles.Set("md-heading-ground", srgb.MustHex("#919c74").Swatch())
	style, err := Of(roles)
	if err != nil {
		t.Fatal(err)
	}
	got := colours(t, style)
	if got["heading.color"] != "#112233" || got["heading.background_color"] != "#919c74" {
		t.Errorf("heading %q on %q", got["heading.color"], got["heading.background_color"])
	}
}

// a style read back and said again is the same style, byte for byte.
func TestReadOfIsIdentity(t *testing.T) {
	first, err := Of(editor())
	if err != nil {
		t.Fatal(err)
	}
	roles, err := Read(first)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Of(roles)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Error("read and said again, the style changed")
	}
}

// with no ink there is nothing for the rest to fall back to.
func TestOfNeedsTheInk(t *testing.T) {
	if _, err := Of(css.New()); err == nil {
		t.Error("a style was made from nothing")
	}
}

// styled is a style as the json says it, decoded.
func styled(t *testing.T, text string) map[string]any {
	t.Helper()
	style := map[string]any{}
	if err := json.Unmarshal([]byte(text), &style); err != nil {
		t.Fatal(err)
	}
	return style
}

// said is a style as json again, so a test can compare two by their text.
func said(t *testing.T, style map[string]any) string {
	t.Helper()
	text, err := json.Marshal(style)
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

// Setting a dotted key makes the objects on the way, replaces a parent that
// is not an object, and leaves every other setting as it was.
func TestSetMakesItsParents(t *testing.T) {
	for name, one := range map[string]struct{ before, key, after string }{
		"the key is there": {`{"a":{"b":"x","c":"y"}}`, "a.b",
			`{"a":{"b":"new","c":"y"}}`},
		"a leaf is new":   {`{"a":{"c":"y"}}`, "a.b", `{"a":{"b":"new","c":"y"}}`},
		"parents are new": {`{"z":"y"}`, "a.b.c", `{"a":{"b":{"c":"new"}},"z":"y"}`},
		"a parent is not an object": {`{"a":"text"}`, "a.b",
			`{"a":{"b":"new"}}`},
		"one part": {`{"a":"old","z":"y"}`, "a", `{"a":"new","z":"y"}`},
	} {
		style := styled(t, one.before)
		set(style, one.key, "new")
		if got := said(t, style); got != one.after {
			t.Errorf("%s: %s, want %s", name, got, one.after)
		}
	}
}

// Removing a dotted key takes that one value out. It makes nothing: a key
// whose parents are missing, or a parent that is not an object, leaves the
// style exactly as it was.
func TestRemoveTakesOneValueAndMakesNothing(t *testing.T) {
	for name, one := range map[string]struct{ before, key, after string }{
		"the key is there": {`{"a":{"b":"x","c":"y"}}`, "a.b", `{"a":{"c":"y"}}`},
		"the leaf is not there": {`{"a":{"c":"y"}}`, "a.b",
			`{"a":{"c":"y"}}`},
		"a parent is not there": {`{"z":"y"}`, "a.b.c", `{"z":"y"}`},
		"a parent halfway":      {`{"a":{"z":"y"}}`, "a.b.c", `{"a":{"z":"y"}}`},
		"a parent is not an object": {`{"a":"text"}`, "a.b",
			`{"a":"text"}`},
		"one part": {`{"a":"old","z":"y"}`, "a", `{"z":"y"}`},
	} {
		style := styled(t, one.before)
		remove(style, one.key)
		if got := said(t, style); got != one.after {
			t.Errorf("%s: %s, want %s", name, got, one.after)
		}
	}
}

// a ground nobody can fall back to is taken out of the style, not left
// in base.json's colours.
func TestUnsetGroundsLeave(t *testing.T) {
	roles := css.New()
	roles.Set("ink", srgb.MustHex("#38241e").Swatch())
	style, err := Of(roles)
	if err != nil {
		t.Fatal(err)
	}
	got := colours(t, style)
	for _, key := range []string{"code.background_color",
		chroma + "background.background_color"} {
		if value, found := got[key]; found {
			t.Errorf("%s is still %s", key, value)
		}
	}
}

// remember is deferred to keep what a panic was.
func remember(cause *any) { *cause = recover() }

// stopped is what the function panicked with, or nil.
func stopped(run func()) (cause any) {
	defer remember(&cause)
	run()
	return nil
}

// readTheBase is a call that hands styleOf the embedded base.
func readTheBase() { styleOf(base) }

// notAStyle is a call that hands styleOf bytes that are no json object.
func notAStyle() { styleOf([]byte("{")) }

// The anti-test: a base that is not json stops the program and names the
// file, and a base that is json is read. The base is embedded, so the
// fault is the program's own and there is no caller to hand it to.
func TestABaseThatIsNotJSONStopsTheProgram(t *testing.T) {
	err, ok := stopped(notAStyle).(error)
	if !ok || !strings.Contains(err.Error(), "base.json") {
		t.Errorf("styleOf({) stopped with %v", err)
	}
	if got := stopped(readTheBase); got != nil {
		t.Errorf("the embedded base stopped the program: %v", got)
	}
}

// The anti-test: Of does not leave its changes in the base. A style made
// for one colourway is not in the next one's, so a second colourway with
// less in it does not inherit the first's colours.
func TestOneStyleDoesNotLeakIntoTheNext(t *testing.T) {
	bare := css.New()
	bare.Set("ink", srgb.MustHex("#38241e").Swatch())
	first, err := Of(bare)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Of(editor()); err != nil {
		t.Fatal(err)
	}
	second, err := Of(bare)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Error("a colourway in between changed the next style")
	}
}
