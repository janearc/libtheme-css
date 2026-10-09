package colourway

import (
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// what a colourway sets, Resolve keeps as it is, and says nothing about.
func TestResolveKeepsWhatIsSet(t *testing.T) {
	source := terminal().Roles
	source.Set("heading", srgb.MustHex("#123456").Swatch())
	resolved := Resolve(source)
	for _, name := range source.Names() {
		want, _ := source.Get(name)
		if got, _ := resolved.Of(name); got != want {
			t.Errorf("%s moved", name)
		}
		if from, said := resolved.From[name]; said {
			t.Errorf("%s set, and said to come from %s", name, from)
		}
	}
}

// each text and state role the colourway does not set comes from the
// sixteen by the table, and says so.
func TestResolveFromTheSixteen(t *testing.T) {
	source := terminal().Roles
	resolved := Resolve(source)
	for name, from := range map[string]string{"heading": "magenta", "link": "blue",
		"code": "green", "string": "yellow", "dim": "bright-black", "err": "red",
		"warn": "yellow", "info": "blue", "hint": "cyan", "ok": "green"} {
		want, _ := source.Get(from)
		if got, _ := resolved.Of(name); got != want || resolved.From[name] != from {
			t.Errorf("%s came from %q", name, resolved.From[name])
		}
	}
}

// another general role comes before the sixteen: gutter and border from
// dim, whether dim is set or itself derived.
func TestResolveFollowsGeneralRoles(t *testing.T) {
	source := terminal().Roles
	derived := Resolve(source)
	brightBlack, _ := source.Get("bright-black")
	source.Set("dim", srgb.MustHex("#445566").Swatch())
	set := Resolve(source)
	dim, _ := source.Get("dim")
	for _, name := range []string{"gutter", "border"} {
		if got, _ := derived.Of(name); got != brightBlack || derived.From[name] != "dim" {
			t.Errorf("%s with dim derived came from %q", name, derived.From[name])
		}
		if got, _ := set.Of(name); got != dim {
			t.Errorf("%s did not follow the dim set", name)
		}
	}
}

// a ground the colourway does not set is its ground mixed toward its ink
// in oklab, by that ground's share; with no ink, the ground itself.
func TestResolveMixesGrounds(t *testing.T) {
	source := terminal().Roles
	resolved := Resolve(source)
	ground, _ := source.Get("ground")
	ink, _ := source.Get("ink")
	for name, share := range mixedToward {
		got, _ := resolved.Of(name)
		if got != ok.Mix.Mix(ground, ink, share) || resolved.From[name] != "ground and ink" {
			t.Errorf("%s is not the ground mixed %.2f toward the ink", name, share)
		}
	}
	if !(mixedToward["cursor-line"] < mixedToward["panel"] &&
		mixedToward["panel"] < mixedToward["surface-1"]) {
		t.Error("the grounds are not mixed a little, more, more again")
	}
	alone := css.New()
	alone.Set("ground", ground)
	if got, _ := Resolve(alone).Of("panel"); got != ground {
		t.Error("with no ink, a ground is not the ground")
	}
}

// with nothing else to come from, text is the ink; with no ink and no
// ground there is nothing to derive, and nothing is.
func TestResolveLastResort(t *testing.T) {
	inkOnly := css.New()
	ink := srgb.MustHex("#38241e").Swatch()
	inkOnly.Set("ink", ink)
	resolved := Resolve(inkOnly)
	if got, _ := resolved.Of("heading"); got != ink || resolved.From["heading"] != "ink" {
		t.Errorf("heading with only an ink came from %q", resolved.From["heading"])
	}
	if names := Resolve(css.New()).Roles.Names(); len(names) != 0 {
		t.Errorf("an empty colourway resolved to %v", names)
	}
}

// claude code's colours the colourway does not set follow the general
// roles by the claude dialect's table, a diff's bands as the ground mixed
// toward ok and err; one it sets, by name or by number, is left alone.
func TestResolveClaude(t *testing.T) {
	source := terminal().Roles
	source.Set("palette-188", srgb.MustHex("#010203").Swatch())
	resolved := Resolve(source)
	ground, _ := source.Get("ground")
	text, _ := source.Get("claude-text")
	if got, _ := resolved.Of("claude-text"); got != text || resolved.From["claude-text"] != "" {
		t.Error("the claude text the colourway sets was moved")
	}
	if _, found := resolved.Of("claude-words"); found {
		t.Error("claude-words was derived though palette-188 is set")
	}
	info, _ := resolved.Of("info")
	if got, _ := resolved.Of("claude-mode"); got != info || resolved.From["claude-mode"] != "info" {
		t.Errorf("claude-mode came from %q", resolved.From["claude-mode"])
	}
	for name, from := range map[string]string{"claude-added": "ok", "claude-removed": "err"} {
		toward, _ := resolved.Of(from)
		if got, _ := resolved.Of(name); got != ok.Mix.Mix(ground, toward, 0.25) ||
			resolved.From[name] != "ground and "+from {
			t.Errorf("%s came from %q", name, resolved.From[name])
		}
	}
	if _, said := resolved.From["black"]; said {
		t.Error("black, the sixteen's own, was derived")
	}
}
