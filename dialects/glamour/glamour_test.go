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
