package css

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// Colourway is what a colourway file holds once read: its roles, every
// custom property that is a hex colour, in file order, and its ramps.
// A ramp comes from two shapes a file uses. A numbered family of roles,
// --sunset-1 to --sunset-6, is an even ramp called sunset. A rule whose
// background is a gradient with hex stops at percentages,
// `.sunset { background: linear-gradient(180deg, #2a0f4d 0%, ...) }`,
// is a positioned ramp called after the selector; where both name the
// same ramp the gradient wins, since it carries the positions. Stops
// that are not plain hex, `transparent` and `rgba(...)` with an alpha,
// are left out: a lamp has no alpha and a sheet has no ground to blend
// with. Ramps mix in oklab, the library's line between two colours,
// which is not what a browser does with the same gradient and is said
// here so nobody is surprised.
type Colourway struct {
	Roles *Sheet
	Ramps map[string]functions.Ramp
	// Order is the ramps' names in the order they were found.
	Order []string
}

var (
	roleRe     = regexp.MustCompile(`--([a-zA-Z0-9_-]+)\s*:\s*(#[0-9a-fA-F]{3,8})\b`)
	ruleRe     = regexp.MustCompile(`([.#]?[a-zA-Z0-9_-]+)\s*\{([^}]*)\}`)
	gradientRe = regexp.MustCompile(`(?:repeating-)?(?:linear|radial)-gradient\(([^;]*)\)`)
	stopRe     = regexp.MustCompile(`(#[0-9a-fA-F]{3,8})\s+(\d+(?:\.\d+)?)%`)
	familyRe   = regexp.MustCompile(`^(.*)-(\d+)$`)
)

// Read parses a colourway's css. It is not a css parser; it reads the
// two shapes a colourway file uses and ignores the rest, which is what
// a reader that will be pointed at hand-written files should do.
func Read(src string) Colourway {
	cw := Colourway{Roles: New(), Ramps: map[string]functions.Ramp{}}
	for _, m := range roleRe.FindAllStringSubmatch(src, -1) {
		if c, err := srgb.FromHex(m[2]); err == nil {
			cw.Roles.Set(m[1], c.Swatch())
		}
	}
	// numbered families first, so a gradient with the same name replaces
	// them below rather than the other way round.
	families := map[string][]struct {
		n    int
		name string
	}{}
	for _, name := range cw.Roles.Names() {
		if m := familyRe.FindStringSubmatch(name); m != nil {
			n, _ := strconv.Atoi(m[2])
			families[m[1]] = append(families[m[1]], struct {
				n    int
				name string
			}{n, name})
		}
	}
	var familyNames []string
	for f := range families {
		familyNames = append(familyNames, f)
	}
	sort.Strings(familyNames)
	for _, f := range familyNames {
		members := families[f]
		if len(members) < 2 {
			continue
		}
		sort.Slice(members, func(i, j int) bool { return members[i].n < members[j].n })
		var stops []swatchAt
		for i, mb := range members {
			c, _ := cw.Roles.Get(mb.name)
			stops = append(stops, swatchAt{float64(i) / float64(len(members)-1), c})
		}
		cw.add(f, stops)
	}
	for _, m := range ruleRe.FindAllStringSubmatch(src, -1) {
		g := gradientRe.FindStringSubmatch(m[2])
		if g == nil {
			continue
		}
		var stops []swatchAt
		for _, st := range stopRe.FindAllStringSubmatch(g[1], -1) {
			c, err := srgb.FromHex(st[1])
			if err != nil {
				continue
			}
			at, _ := strconv.ParseFloat(st[2], 64)
			stops = append(stops, swatchAt{at / 100, c.Swatch()})
		}
		if len(stops) == 0 {
			continue
		}
		name := strings.TrimLeft(m[1], ".#")
		if _, seen := cw.Ramps[name]; seen {
			cw.Ramps[name] = ramp(stops)
			continue
		}
		cw.add(name, stops)
	}
	return cw
}

type swatchAt struct {
	at float64
	c  swatch.Swatch
}

func (cw *Colourway) add(name string, stops []swatchAt) {
	cw.Ramps[name] = ramp(stops)
	cw.Order = append(cw.Order, name)
}

func ramp(stops []swatchAt) functions.Ramp {
	out := make([]functions.Stop, 0, len(stops))
	for _, s := range stops {
		out = append(out, functions.Stop{At: s.at, Swatch: s.c})
	}
	return functions.New(ok.Mix, out...)
}
