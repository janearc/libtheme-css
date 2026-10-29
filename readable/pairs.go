package readable

import (
	"math"
	"slices"

	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// fullCircle is the degrees round the wheel. It is defined by the circle.
const fullCircle = 360

// pairs holds each pair a highlight group draws on its own, text and
// ground both set by the group, to the floor and the hue rule. A search
// match is the ground colour on the string colour, which no surface check
// sees.
//
// Furniture is left alone, and so is a pair that is a prose or state
// colour on a surface, which is already held to more.
func (held gate) pairs() []violation {
	var found []violation
	seen := map[string]bool{}
	for _, group := range nvim.Groups() {
		ink, ground := effective(group)
		pair := ink + " on " + ground
		if seen[pair] || slices.Contains(furniture, ink) ||
			held.onASurface(ink, ground) {
			continue
		}
		seen[pair] = true
		role := group.Name + " draws " + pair
		text := held.scheme.Colours[ink]
		behind := held.scheme.Colours[ground]
		found = append(found, held.dim(role, "", text, behind)...)
		found = append(found, held.hue(role, "", text, behind)...)
	}
	return found
}

// effective is the names of the text and the ground a group draws with. A
// colour the group does not set is the editor's own: the ink, or the
// ground.
func effective(group nvim.Group) (ink, ground string) {
	ink, ground = group.Fg, group.Bg
	if ink == "" {
		ink = "ink"
	}
	if ground == "" {
		ground = "bg"
	}
	return ink, ground
}

// onASurface is whether a pair is a prose or state colour on a surface.
func (held gate) onASurface(ink, ground string) bool {
	text := slices.Contains(prose, ink) || slices.Contains(states, ink)
	for _, each := range surfaces {
		if each.name == ground {
			return text
		}
	}
	return false
}

// lamps is a swatch on the lamps.
func lamps(value swatch.Swatch) srgb.RGB {
	lit, _ := srgb.FromSwatch(value)
	return lit
}

// hex is a swatch as the file would write it.
func hex(value swatch.Swatch) string { return lamps(value).Hex() }

// hueApart is the distance round the wheel between two hues in degrees, up
// to the half circle.
func hueApart(first, second float64) float64 {
	apart := math.Abs(first - second)
	if apart > halfCircle {
		apart = fullCircle - apart
	}
	return apart
}
