package css

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// Colourway is what a colourway file holds once read. Roles are every custom
// property that is a hex colour, in file order. Ramps are numbered families
// of roles: --sunset-1 to --sunset-6 is an even ramp called sunset. Each
// family needs two members at least.
//
// That is the whole grammar, and it is ours. It reads a sheet this library
// wrote, or a colourway written by hand in the same shape. Nothing else in
// the file is read. That includes selectors, gradients and the rest of css.
//
// Hex colours are complete (#rgb or #rrggbb) and outside any comment.
// Custom properties that start like a hex colour and are not one are not
// read, and Skipped says so.
//
// Properties that do not start like a hex colour are not read either. Named
// lists the ones a caller says are roles, so a role written in oklch or rgb
// can be refused and not lost.
//
// This is a container, not a parser. Files that want a ramp read promote
// the ramp's stops to roles. Ramps mix in oklab, the library's line
// between two colours.
type Colourway struct {
	Roles *Sheet
	Ramps map[string]functions.Ramp
	// Order is the ramps' names in the order they were found.
	Order   []string
	skipped []string
	others  []property
}

// property is a custom property the reader did not take as a colour. It has
// a name and a line. The line is the one the name is on. The value is as
// the file wrote it.
type property struct {
	name  string
	line  int
	value string
}

// Skipped lists the custom properties that start like a hex colour and were
// not read. Each has its line and its value as the file wrote it.
func (cw Colourway) Skipped() []string {
	return append([]string(nil), cw.skipped...)
}

// Named lists the custom properties that do not start like a hex colour, as
// Skipped does for the incomplete ones. The caller says which names it reads
// and only those are listed. Roles in oklch or rgb are listed. Fonts are
// not, unless the caller reads a role of that name.
func (cw Colourway) Named(reads func(string) bool) []string {
	var said []string
	for _, each := range cw.others {
		if reads(each.name) {
			said = append(said,
				notAColour(each.line, each.name, each.value))
		}
	}
	return said
}

// notAColour is what Skipped and Named say of a property.
func notAColour(line int, name, written string) string {
	return fmt.Sprintf("line %d: --%s: %s is not a #rgb or #rrggbb",
		line, name, written)
}

var (
	// declarationRe finds a custom property whose value starts like a hex
	// colour, and takes everything to the end of the declaration.
	declarationRe = regexp.MustCompile(
		`--([a-zA-Z0-9_-]+)\s*:\s*(#[^;}]*)`)
	// otherRe finds any custom property. It must follow the start of the
	// text or a space or a semicolon or a brace. Names inside other words
	// are not properties. The value group is the rest of the declaration.
	otherRe = regexp.MustCompile(
		`(?:^|[\s;{])--([a-zA-Z0-9_-]+)\s*:([^;}]*)`)
	// colourRe is a complete hex colour.
	colourRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	familyRe = regexp.MustCompile(`^(.*)-(\d+)$`)
	// importantRe is the flag that may follow a colour. The bang and the
	// word may be in any case, with space between them.
	importantRe = regexp.MustCompile(`(?i)\s*!\s*important$`)
)

// opens and closes are a css comment's marks.
const (
	opens  = "/*"
	closes = "*/"
)

// blank turns a comment into spaces, one for each byte. Line ends stay line
// ends, so line numbers hold. Each place in the result is the same place in
// the source.
func blank(comment string) string {
	spaces := []byte(comment)
	for at, each := range spaces {
		if each != '\n' {
			spaces[at] = ' '
		}
	}
	return string(spaces)
}

// uncommented is the source with every comment blanked out. It is as long
// as the source. Unclosed comments take the rest of the file.
func uncommented(src string) string {
	var text strings.Builder
	for rest := src; ; {
		before, inside, found := strings.Cut(rest, opens)
		text.WriteString(before)
		if !found {
			return text.String()
		}
		comment, after, closed := strings.Cut(inside, closes)
		if closed {
			comment += closes
		}
		text.WriteString(blank(opens + comment))
		rest = after
	}
}

// complaint is what Skipped says of a declaration. It names the line the
// name is on. It quotes the value as the file wrote it, comments too.
// Searching the file for the quote finds the place. The text is the source
// with its comments blanked. The at is where the regexp matched in it.
func complaint(src, text string, at []int) string {
	name := text[at[nameFrom]:at[nameTo]]
	line := strings.Count(text[:at[matchFrom]], "\n") + lineOne
	written := strings.TrimSpace(src[at[valueFrom]:at[valueTo]])
	return notAColour(line, name, written)
}

// read takes the roles from a source, and notes the colours it could not.
func (cw *Colourway) read(src string) {
	text := uncommented(src)
	found := declarationRe.FindAllStringSubmatchIndex(text, everyMatch)
	for _, at := range found {
		name := text[at[nameFrom]:at[nameTo]]
		value := text[at[valueFrom]:at[valueTo]]
		value = strings.TrimSpace(value)
		colour := importantRe.ReplaceAllString(value, "")
		if !colourRe.MatchString(colour) {
			said := complaint(src, text, at)
			cw.skipped = append(cw.skipped, said)
			continue
		}
		cw.Roles.Set(name, srgb.MustHex(colour).Swatch())
	}
	cw.others = othersIn(src, text)
}

// othersIn is the custom properties of a source whose value does not start
// like a hex colour. The text is the source with its comments blanked.
func othersIn(src, text string) []property {
	var found []property
	matches := otherRe.FindAllStringSubmatchIndex(text, everyMatch)
	for _, at := range matches {
		written := strings.TrimSpace(src[at[valueFrom]:at[valueTo]])
		if strings.HasPrefix(written, "#") {
			continue
		}
		line := strings.Count(text[:at[nameFrom]], "\n") + lineOne
		found = append(found, property{
			text[at[nameFrom]:at[nameTo]], line, written,
		})
	}
	return found
}

// Where the parts of a declaration sit in what the regexp reports: it
// gives the start and end of the whole match, then of each group. Each is
// defined by the regexp package.
const (
	// matchFrom is where the match starts; defined by the regexp package.
	matchFrom = 0
	// nameFrom is where the name group starts; defined by the regexp
	// package.
	nameFrom = 2
	// nameTo is where the name group ends; defined by the regexp package.
	nameTo = 3
	// valueFrom is where the value group starts; defined by the regexp
	// package.
	valueFrom = 4
	// valueTo is where the value group ends; defined by the regexp package.
	valueTo = 5
)

// everyMatch asks regexp for every match, not a count of them. It is
// defined by the regexp package: a negative count is all.
const everyMatch = -1

// lineOne is the number of the first line of a file. It is defined by
// counting lines from one.
const lineOne = 1

// Read reads a colourway. Hex custom properties become roles. Numbered
// families become ramps. It is not a css parser and reads nothing else.
func Read(src string) Colourway {
	cw := Colourway{Roles: New(), Ramps: map[string]functions.Ramp{}}
	cw.read(src)
	families := map[string][]member{}
	var order []string
	for _, name := range cw.Roles.Names() {
		if m := familyRe.FindStringSubmatch(name); m != nil {
			n, _ := strconv.Atoi(m[2])
			if _, seen := families[m[1]]; !seen {
				order = append(order, m[1])
			}
			families[m[1]] = append(families[m[1]], member{n, name})
		}
	}
	for _, f := range order {
		members := families[f]
		if len(members) < 2 {
			continue
		}
		slices.SortStableFunc(members, byNumber)
		var stops []swatchAt
		for i, mb := range members {
			c, _ := cw.Roles.Get(mb.name)
			stops = append(stops,
				swatchAt{float64(i) / float64(len(members)-1),
					c})
		}
		cw.add(f, stops)
	}
	return cw
}

// member is a role of a numbered family: the number in its name, and the
// name.
type member struct {
	n    int
	name string
}

// byNumber orders members by their numbers. The sort that uses it is
// stable, so members with one number stay in the order the file had them.
func byNumber(first, second member) int {
	return cmp.Compare(first.n, second.n)
}

type swatchAt struct {
	at float64
	c  swatch.Swatch
}

// add records a ramp under its name, keeping the order it was found in.
func (cw *Colourway) add(name string, stops []swatchAt) {
	cw.Ramps[name] = ramp(stops)
	cw.Order = append(cw.Order, name)
}

// ramp is stops at positions as a ramp mixed in oklab.
func ramp(stops []swatchAt) functions.Ramp {
	out := make([]functions.Stop, 0, len(stops))
	for _, s := range stops {
		out = append(out, functions.Stop{At: s.at, Swatch: s.c})
	}
	return functions.New(ok.Mix, out...)
}
