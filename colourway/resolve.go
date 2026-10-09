package colourway

import (
	"fmt"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
)

// the general vocabulary: roles every dialect reads from. They are more
// general than any one program's. Program roles like nvim-ink or md-heading
// stay overrides. The dialect reads them first.
var (
	Grounds = []string{"ground", "surface-1", "panel", "cursor-line"}
	Text    = []string{"ink", "dim", "heading", "link", "code", "string",
		"gutter", "border"}
	States = []string{"err", "warn", "info", "hint", "ok"}
)

// derivedFrom is where a text or state role comes from when a colourway
// does not set it. Each role names one source: another general role, or one
// of the terminal's sixteen, which every terminal colourway has. If that
// source is not set either, the role falls back to the ink.
//
// Classic terminal tables disagree about heading, link and code. These are
// picked.
var derivedFrom = map[string][]string{
	"dim": {"bright-black"}, "gutter": {"dim"}, "border": {"dim"},
	"heading": {"magenta"}, "link": {"blue"}, "code": {"green"},
	"string": {"yellow"}, "err": {"red"}, "warn": {"yellow"},
	"info": {"blue"}, "hint": {"cyan"}, "ok": {"green"},
}

// mixedToward is how far a ground the colourway does not set is mixed from
// the ground toward the ink, in oklab. The numbers are picked, from the
// middle of what hand-tuned colourways do.
var mixedToward = map[string]float64{
	"cursor-line": 0.06, "panel": 0.07, "surface-1": 0.13,
}

// Resolved is a colourway's roles with the general vocabulary filled in.
// From says what each filled role came from.
type Resolved struct {
	Roles *css.Sheet
	From  map[string]string
}

// Resolve makes a colourway's roles whole for the programs above the
// terminal. Every role the colourway sets is kept as it is.
//
// Each general role it does not set is derived. Text and state roles come
// from another general role or one of the sixteen, and failing those from
// the ink. Grounds are the ground mixed toward the ink, or the ground itself
// when there is no ink.
//
// Then claude code's colours are filled in from the general roles, by the
// claude dialect's table. A colour the colourway sets by its number is left
// alone. A role is left out only when there is nothing to derive it from.
func Resolve(roles *css.Sheet) Resolved {
	resolved := Resolved{Roles: css.New(), From: map[string]string{}}
	for _, name := range roles.Names() {
		value, _ := roles.Get(name)
		resolved.Roles.Set(name, value)
	}
	for _, name := range append(append([]string{}, Text...), States...) {
		resolved.text(name)
	}
	for _, name := range Grounds {
		resolved.ground(name)
	}
	for _, slot := range claude.Slots {
		number := fmt.Sprintf("palette-%d", slot.Number)
		if _, numbered := resolved.Roles.Get(number); !numbered {
			resolved.claude(slot.Role, slot.From, slot.Toward)
		}
	}
	return resolved
}

// claude gives one of claude code's colours a colour if it has none. It
// takes the general role the colour follows. For a diff's bands it takes the
// ground mixed that far toward that role, so the bands stay grounds and the
// code on them reads.
func (resolved Resolved) claude(name, from string, toward float64) {
	if _, found := resolved.Roles.Get(name); found || from == "" {
		return
	}
	value, found := resolved.Roles.Get(from)
	if !found {
		return
	}
	if toward > 0 {
		ground, hasGround := resolved.Roles.Get("ground")
		if !hasGround {
			return
		}
		value = ok.Mix.Mix(ground, value, toward)
		from = "ground and " + from
	}
	resolved.Roles.Set(name, value)
	resolved.From[name] = from
}

// text gives a text or state role a colour if it has none. It takes the
// first of its sources that has one, and the ink if none does.
func (resolved Resolved) text(name string) {
	if _, found := resolved.Roles.Get(name); found {
		return
	}
	for _, source := range append(derivedFrom[name], "ink") {
		if value, found := resolved.Roles.Get(source); found {
			resolved.Roles.Set(name, value)
			resolved.From[name] = source
			return
		}
	}
}

// ground gives a ground a colour if it has none. It is the ground mixed
// toward the ink by its share, or the ground itself when there is no ink.
func (resolved Resolved) ground(name string) {
	if _, found := resolved.Roles.Get(name); found {
		return
	}
	ground, hasGround := resolved.Roles.Get("ground")
	if !hasGround {
		return
	}
	ink, hasInk := resolved.Roles.Get("ink")
	if !hasInk {
		resolved.Roles.Set(name, ground)
		resolved.From[name] = "ground"
		return
	}
	resolved.Roles.Set(name, ok.Mix.Mix(ground, ink, mixedToward[name]))
	resolved.From[name] = "ground and ink"
}

// Of is a role's colour once resolved, and whether there is one.
func (resolved Resolved) Of(name string) (swatch.Swatch, bool) {
	return resolved.Roles.Get(name)
}
