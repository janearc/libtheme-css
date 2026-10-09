package colourway

import (
	"fmt"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
)

// the general vocabulary: roles every dialect reads from, more general
// than any one program's. A program's own name for a role, like nvim-ink
// or md-heading, stays an override, and the dialect reads it first.
var (
	Grounds = []string{"ground", "surface-1", "panel", "cursor-line"}
	Text    = []string{"ink", "dim", "heading", "link", "code", "string",
		"gutter", "border"}
	States = []string{"err", "warn", "info", "hint", "ok"}
)

// derivedFrom is where a text or state role comes from when a colourway
// does not set it, in order: another general role, then the terminal's
// sixteen, which every terminal colourway has.
//
// It is the table drafted in the vocabulary proposal, taken 2026-10-05;
// the classic tables disagree about heading, link and code.
var derivedFrom = map[string][]string{
	"dim": {"bright-black"}, "gutter": {"dim"}, "border": {"dim"},
	"heading": {"magenta"}, "link": {"blue"}, "code": {"green"},
	"string": {"yellow"}, "err": {"red"}, "warn": {"yellow"},
	"info": {"blue"}, "hint": {"cyan"}, "ok": {"green"},
}

// mixedToward is how far a ground the colourway does not set is mixed
// from its ground toward its ink, in oklab: about the middle of what her
// own colourways do (twilight, twilight-deep, vaporwave, vaporwave-dusk,
// aqua and vim-dusk, measured 2026-10-05).
var mixedToward = map[string]float64{
	"cursor-line": 0.06, "panel": 0.07, "surface-1": 0.13,
}

// Resolved is a colourway's roles with the general vocabulary filled in,
// and, for each role it filled, what it came from.
type Resolved struct {
	Roles *css.Sheet
	From  map[string]string
}

// Resolve is a colourway's roles made whole for the programs above the
// terminal. Every role it sets is kept as it is.
//
// Each general role it does not set is derived: from another general role
// or the sixteen, a ground by mixing the ground toward the ink, and
// failing those the ink for text and the ground for a ground.
//
// Then claude code's colours, from the general roles by the claude
// dialect's table, unless the colourway sets one by its number. A role is
// left out only when there is nothing at all to derive it from.
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

// claude gives one of claude code's colours a colour if it has none: the
// general role it follows, or for a diff's bands the ground mixed that
// far toward it, so the bands stay grounds and the code on them reads.
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

// text gives a text or state role a colour if it has none: the first of
// its sources that has one, else the ink.
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

// ground gives a ground a colour if it has none: the ground mixed toward
// the ink by its share, or the ground itself when there is no ink.
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
