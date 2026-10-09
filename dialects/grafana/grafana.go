// Package grafana is the grafana dialect: how grafana names a colour.
//
// Grafana does not take a theme. It names colours in a vocabulary of its
// own: a shade in front of a hue. Everywhere else it asks for a hex in a
// fixed-colour field.
//
// The five names of a hue are dark, semi-dark, the bare hue, light and
// super-light. The dialect reads them as distances along the line from
// the colour towards black or towards white. The mix is in oklab, which
// keeps the hue. A lighter green stays green.
//
// Moving the lightness alone does not keep the hue. Past the edge of the
// gamut, the only way to keep going is to drop the chroma. Then three of
// the five names come out white.
//
// The dialect translates both ways. A swatch becomes the five shades, and
// a name becomes the swatch it stands for.
package grafana

import (
	"strings"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// Shades are grafana's names for one hue, darkest first. The empty string
// is the bare hue.
var Shades = []string{"dark", "semi-dark", "", "light", "super-light"}

// step is how far along the line each shade sits. Negative is towards
// black and positive is towards white. It is a fraction of the whole
// distance. A quarter reads as one step of a palette on screen. Half is
// as far as a shade goes before it stops being the same colour.
var step = map[string]float64{
	"dark": -0.50, "semi-dark": -0.25, "": 0, "light": 0.25,
	"super-light": 0.50,
}

// Of is the five shades grafana expects of one hue, keyed by shade name.
func Of(c swatch.Swatch) map[string]swatch.Swatch {
	out := map[string]swatch.Swatch{}
	for _, s := range Shades {
		out[s] = Shade(c, s)
	}
	return out
}

// Shade is one named shade of a colour. It moves that far along the line
// towards black or towards white, mixed in oklab. Then it brings the
// result inside the gamut, because grafana cannot show a hex outside it.
// The bare hue, or an unknown shade, returns the colour unchanged.
func Shade(c swatch.Swatch, shade string) swatch.Swatch {
	d, known := step[shade]
	if !known || d == 0 {
		return c
	}
	end := swatch.White
	if d < 0 {
		end, d = swatch.Black, -d
	}
	line := functions.Even(ok.Mix, c, end)
	fit, _ := ok.Fit(ok.FromSwatch(line.At(d)).Polar(), InGamut)
	return fit.Rect().Swatch()
}

// InGamut reports whether sRGB can name the swatch. That is the only
// question grafana's hex field can answer.
func InGamut(s swatch.Swatch) bool {
	_, in := srgb.FromSwatch(s)
	return in
}

// Name is grafana's name for a hue at a shade: "semi-dark-blue", or
// "blue" for the bare hue.
func Name(shade, hue string) string {
	if shade == "" {
		return hue
	}
	return shade + "-" + hue
}

// Split reads one of grafana's names back into its shade and its hue. A
// name with no shade in front of it is the bare hue. Its shade is empty.
func Split(name string) (shade, hue string) {
	for _, s := range Shades {
		if s == "" {
			continue
		}
		if strings.HasPrefix(name, s+"-") {
			return s, strings.TrimPrefix(name, s+"-")
		}
	}
	return "", name
}

// Palette is every name grafana can use for the hues it is given. Each hue
// has five shades, as hex, ready for a dashboard's fixed-colour fields.
// The keys are grafana's names. No other names are invented.
func Palette(hues map[string]swatch.Swatch) map[string]string {
	out := map[string]string{}
	for hue, c := range hues {
		for shade, s := range Of(c) {
			rgb, _ := srgb.FromSwatch(s)
			out[Name(shade, hue)] = rgb.Hex()
		}
	}
	return out
}

// Lookup is the swatch one of grafana's names stands for, given the hues
// it was built from. If the name's hue is not bound, the name is not ours.
// Lookup then returns false instead of guessing.
func Lookup(name string, hues map[string]swatch.Swatch) (swatch.Swatch, bool) {
	shade, hue := Split(name)
	c, bound := hues[hue]
	if !bound {
		return swatch.Swatch{}, false
	}
	return Shade(c, shade), true
}
