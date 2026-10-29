package readable

import (
	"fmt"

	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// tooDim is the rule a prose colour breaks under the band.
const tooDim = "too dim to read at length"

// prose are the scheme's colours that carry text read at length.
var prose = []string{"ink", "dim", "heading", "link", "code", "string"}

// states are the scheme's colours that are scanned, not read.
var states = []string{"err", "warn", "info", "hint"}

// furniture are the colours you look past: line numbers and borders. The
// policy has no rule for them.
var furniture = []string{"gutter", "border"}

// surface is a ground text can land on: its name in the scheme, and what
// to call it.
type surface struct {
	name, says string
}

// surfaces are the grounds text can land on. Text lands on the ground; on
// the cursor's line; on the selection, which draws over text; and on a
// float or the status line.
var surfaces = []surface{
	{"bg", "ground"},
	{"cursorln", "cursor line"},
	{"visual", "selection"},
	{"panel", "panel"},
}

// Report is what a check found. It says nothing of colours the policy has
// no rule for: the gutter and the borders are furniture you look past, and
// the terminal's sixteen are not text.
type Report struct {
	found []violation
}

// OK is whether the scheme is inside the reader's range everywhere it was
// held to it.
func (report Report) OK() bool { return len(report.found) == nothing }

// Lines are the failures, one line each, in a stable order. Each says which
// colour on which surface broke which rule, with the numbers.
func (report Report) Lines() []string {
	var lines []string
	for _, each := range report.found {
		lines = append(lines, each.String())
	}
	return lines
}

// violation is one failure, with the numbers that make it one.
type violation struct {
	scheme, role, surface, rule, detail string
}

// String says what is wrong in one line a person can act on.
func (failure violation) String() string {
	where := failure.role
	if failure.surface != "" {
		where += " on the " + failure.surface
	}
	return fmt.Sprintf("%s: %s: %s (%s)", failure.scheme, where,
		failure.rule, failure.detail)
}

// gate is a scheme held to a policy.
type gate struct {
	scheme nvim.Scheme
	policy Policy
}

// Check holds a scheme to a policy and returns every failure. Each prose
// colour is held to the band on every surface, each state colour to the
// floor on every surface, and each pair a highlight group draws on its
// own, such as the search match, to the floor.
//
// A colour the scheme lacks is a failure, not a skip.
func Check(scheme nvim.Scheme, policy Policy) Report {
	held := gate{scheme: scheme, policy: policy}
	found := held.missing()
	if len(found) > nothing {
		return Report{found}
	}
	for _, each := range surfaces {
		for _, role := range prose {
			found = append(found, held.band(role, each)...)
		}
		for _, role := range states {
			found = append(found, held.floor(role, each)...)
		}
	}
	return Report{append(found, held.pairs()...)}
}

// missing is a failure for each colour the checks read that the scheme
// does not have.
func (held gate) missing() []violation {
	var found []violation
	for _, entry := range nvim.Roles {
		if _, has := held.scheme.Colours[entry.Name]; !has {
			found = append(found, held.fail(entry.Name, "",
				"missing colour", "the scheme does not set it"))
		}
	}
	return found
}

// band holds a prose colour to the band on a surface, and to the hue rule.
func (held gate) band(role string, on surface) []violation {
	ink := held.scheme.Colours[role]
	ground := held.scheme.Colours[on.name]
	got := nvim.Contrast(ink, ground)
	found := held.hue(role, on.says, ink, ground)
	switch {
	case got < held.policy.low:
		detail := held.start(ink, got)
		found = append(found, held.fail(role, on.says, tooDim, detail))
	case got > held.policy.high:
		detail := held.end(ink, got)
		found = append(found, held.fail(role, on.says, "glare", detail))
	}
	return found
}

// start says an ink is under the start of the band.
func (held gate) start(ink swatch.Swatch, got float64) string {
	return fmt.Sprintf("%s is %.2f:1, the band starts at %g:1",
		hex(ink), got, held.policy.low)
}

// end says an ink is over the end of the band.
func (held gate) end(ink swatch.Swatch, got float64) string {
	return fmt.Sprintf("%s is %.2f:1, the band ends at %g:1",
		hex(ink), got, held.policy.high)
}

// floor holds a state colour to the floor on a surface, and to the hue
// rule.
func (held gate) floor(role string, on surface) []violation {
	ink := held.scheme.Colours[role]
	ground := held.scheme.Colours[on.name]
	found := held.dim(role, on.says, ink, ground)
	return append(found, held.hue(role, on.says, ink, ground)...)
}

// dim is a failure if ink on ground is under the state floor.
func (held gate) dim(role, says string, ink, ground swatch.Swatch,
) []violation {
	got := nvim.Contrast(ink, ground)
	if got >= held.policy.stateMin {
		return nil
	}
	detail := fmt.Sprintf("%s is %.2f:1, the floor is %g:1", hex(ink), got,
		held.policy.stateMin)
	return []violation{held.fail(role, says, "too dim to scan", detail)}
}

// hue is a failure if the ground has colour in it and the ink is not far
// enough round the wheel from it: an edge of brightness and no edge of
// colour. A ground below the neutral saturation passes, because the ink's
// own colour carries the edge.
func (held gate) hue(role, says string, ink, ground swatch.Swatch,
) []violation {
	behind := lamps(ground).HSL()
	if behind.S < held.policy.neutral {
		return nil
	}
	apart := hueApart(lamps(ink).HSL().H, behind.H)
	if apart >= held.policy.hueApart {
		return nil
	}
	detail := fmt.Sprintf("%.0f degrees from the surface at saturation "+
		"%.2f; a brightness edge and nothing else", apart, behind.S)
	return []violation{held.fail(role, says, "no colour edge", detail)}
}

// fail is a violation of a rule by a role on a surface.
func (held gate) fail(role, says, rule, detail string) violation {
	return violation{scheme: held.scheme.Name, role: role, surface: says,
		rule: rule, detail: detail}
}
