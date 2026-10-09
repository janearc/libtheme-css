// Package nvim is the neovim dialect: a colour scheme said from a
// colourway's roles, and read back into them.
//
// A scheme is a table of colours by role, the highlight groups that use
// them, and the sixteen for neovim's own terminal, so a shell inside vim
// draws like the terminal around it.
//
// The groups are fixed, in groups.lua; only the colours come from the
// colourway. The sixteen are the terminal's, by the names the ghostty
// dialect gives them.
//
// A colourway can give the editor a colour of its own: nvim-ink wins
// over ink here and nowhere else, for an editor whose ink is a little
// brighter than its terminal's, or whose terminal keeps its own black.
package nvim

import (
	_ "embed"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

//go:embed groups.lua
var groups string

// Roles are the scheme's colours in the order it writes them: the name
// the highlight groups use, the role a colourway gives it, and what it
// is for.
var Roles = []struct{ Name, Role, Use string }{
	{"bg", "ground", "the ground"},
	{"ink", "ink", "body prose"},
	{"dim", "dim", "comments"},
	{"heading", "heading", "headings and keywords"},
	{"link", "link", "links and functions"},
	{"code", "code", "code and types"},
	{"string", "string", "strings and numbers"},
	{"gutter", "gutter", "line numbers, furniture you look past"},
	{"cursorln", "cursor-line", "the cursor's line"},
	{"visual", "surface-1", "the selection"},
	{"panel", "panel", "floats and the status line"},
	{"border", "border", "borders and rules"},
	{"err", "err", "errors"},
	{"warn", "warn", "warnings"},
	{"info", "info", "notes"},
	{"hint", "hint", "hints"},
}

// Scheme is a colour scheme: its name, whether its ground is the light
// one (lighter than its ink), its colours by the names the groups use,
// and the sixteen, which a scheme read from a file may not have.
type Scheme struct {
	Name     string
	Light    bool
	Colours  map[string]swatch.Swatch
	Terminal []swatch.Swatch
}

// Of is the scheme a colourway's roles describe. Every role is required,
// since a group with no colour falls back to somebody else's.
func Of(name string, roles *css.Sheet) (Scheme, error) {
	scheme := Scheme{Name: name, Colours: map[string]swatch.Swatch{},
		Terminal: make([]swatch.Swatch, 16)}
	for _, entry := range Roles {
		value, found := pick(roles, entry.Role)
		if !found {
			return Scheme{}, missing(entry.Role,
				"the scheme's "+entry.Name)
		}
		scheme.Colours[entry.Name] = snap(value)
	}
	for number, role := range ghostty.Ansi {
		value, found := pick(roles, role)
		if !found {
			return Scheme{}, missing(role, "the terminal")
		}
		scheme.Terminal[number] = snap(value)
	}
	scheme.Light = lighter(scheme.Colours)
	return scheme, nil
}

// snap is a colour as the file will hold it, on the lamps' byte grid,
// so the contrast written beside it is the contrast of what is written,
// and a scheme read back says the same.
func snap(value swatch.Swatch) swatch.Swatch {
	snapped, _ := parse(ghostty.Hex(value))
	return snapped
}

// pick is a role's colour for the editor: its own, nvim-role, if the
// colourway gives it one, else the role's.
func pick(roles *css.Sheet, role string) (swatch.Swatch, bool) {
	if value, found := roles.Get("nvim-" + role); found {
		return value, true
	}
	return roles.Get(role)
}

// missing is the error for a role the scheme cannot do without.
func missing(role, needs string) error {
	return fmt.Errorf("nvim: %s is missing; %s needs it", role, needs)
}

// lighter is whether a scheme's ground is the light one: lighter than
// its ink. a fixed threshold would call a mid-tone ground dark under ink
// darker still.
func lighter(colours map[string]swatch.Swatch) bool {
	return luminance(colours["bg"]) > luminance(colours["ink"])
}

// Roles is the scheme said back as a colourway's roles: its colours by
// role, then the sixteen by name, if it has them.
func (scheme Scheme) Roles() *css.Sheet {
	roles := css.New()
	for _, entry := range Roles {
		if value, found := scheme.Colours[entry.Name]; found {
			roles.Set(entry.Role, value)
		}
	}
	for number, value := range scheme.Terminal {
		roles.Set(ghostty.Ansi[number], value)
	}
	return roles
}

var (
	// tableEntry is one line of the colour table: `  name = "#rrggbb",`.
	tableEntry = regexp.MustCompile(`^\s*(\w+)\s*=\s*"(#[0-9a-fA-F]{6})"`)
	// hexes are the quoted colours on a line of the terminal's table.
	hexes = regexp.MustCompile(`"(#[0-9a-fA-F]{6})"`)
	// named is the scheme's name, as it sets it.
	named = regexp.MustCompile(`^\s*vim\.g\.colors_name\s*=\s*"([^"]+)"`)
)

// Read is a scheme file's colours: the `local c = {` table by name, and
// the `local term = {` table's sixteen. Everything else, the groups
// among it, is passed over, so a scheme written by hand in the same
// shape reads as well as one this dialect wrote.
func Read(source io.Reader) (Scheme, error) {
	raw, err := io.ReadAll(source)
	if err != nil {
		return Scheme{}, err
	}
	scheme := Scheme{Colours: map[string]swatch.Swatch{}}
	place := reading{}
	for number, text := range strings.Split(string(raw), "\n") {
		if err := scheme.take(text, &place); err != nil {
			return Scheme{}, fmt.Errorf("nvim: line %d: %w",
				number+1, err)
		}
	}
	if len(scheme.Colours) == 0 {
		return Scheme{}, fmt.Errorf("nvim: no colour table")
	}
	if count := len(scheme.Terminal); count != 0 && count != 16 {
		return Scheme{}, fmt.Errorf("nvim: %d of the sixteen", count)
	}
	_, haveGround := scheme.Colours["bg"]
	_, haveInk := scheme.Colours["ink"]
	scheme.Light = haveGround && haveInk && lighter(scheme.Colours)
	return scheme, nil
}

// reading is where Read is in a scheme file: the table it is inside.
type reading struct {
	table string
}

// take reads one line of a scheme file into the scheme. Comments are
// passed over, whole lines and the ends of lines both, so a colour named
// in a note is not taken for one in the table.
func (scheme *Scheme) take(text string, place *reading) error {
	if code, _, found := strings.Cut(text, "--"); found {
		text = code
	}
	trimmed := strings.TrimSpace(text)
	switch {
	case trimmed == "":
		return nil
	case named.MatchString(text):
		scheme.Name = named.FindStringSubmatch(text)[1]
	case trimmed == "local c = {", trimmed == "local term = {":
		place.table = trimmed
	case strings.HasPrefix(trimmed, "}"):
		place.table = ""
	case place.table == "local c = {":
		return scheme.takeColour(text)
	case place.table == "local term = {":
		return scheme.takeTerminal(text, place)
	}
	return nil
}

// takeColour reads one entry of the colour table.
func (scheme *Scheme) takeColour(text string) error {
	found := tableEntry.FindStringSubmatch(text)
	if found == nil {
		return nil
	}
	value, err := parse(found[2])
	scheme.Colours[found[1]] = value
	return err
}

// takeTerminal reads the sixteen's colours on one line of their table.
func (scheme *Scheme) takeTerminal(text string, place *reading) error {
	for _, found := range hexes.FindAllStringSubmatch(text, -1) {
		if len(scheme.Terminal) == 16 {
			return fmt.Errorf("more than sixteen")
		}
		value, err := parse(found[1])
		if err != nil {
			return err
		}
		scheme.Terminal = append(scheme.Terminal, value)
	}
	return nil
}

// Group is one highlight group as groups.lua sets it: its name, the
// scheme's names for the colours of its letters and its ground, and its
// styles (bold, italic, underline, strikethrough).
type Group struct {
	Name, Fg, Bg string
	Styles       []string
}

// groupLine is one group in groups.lua: g(0, "Name", { key = value, ... }).
var groupLine = regexp.MustCompile(`^g\(0,\s*"([^"]+)",\s*\{([^}]*)\}\)`)

// Groups are the highlight groups every scheme sets, as groups.lua says
// them, so another dialect can say the same groups in its own words.
func Groups() []Group {
	found := []Group{}
	for line := range strings.SplitSeq(groups, "\n") {
		match := groupLine.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			continue
		}
		group := Group{Name: match[1]}
		for field := range strings.SplitSeq(match[2], ",") {
			key, value, set := strings.Cut(field, "=")
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			switch {
			case !set:
			case key == "fg":
				group.Fg = strings.TrimPrefix(value, "c.")
			case key == "bg":
				group.Bg = strings.TrimPrefix(value, "c.")
			case value == "true":
				group.Styles = append(group.Styles, key)
			}
		}
		found = append(found, group)
	}
	return found
}

// preamble is what every scheme says before its colours.
const preamble = `local g = vim.api.nvim_set_hl

vim.cmd("highlight clear")
if vim.fn.exists("syntax_on") == 1 then vim.cmd("syntax reset") end
`

// terminalHead and terminalTail go round the table of the sixteen.
const (
	terminalHead = `-- a shell inside vim (:terminal) draws in these
-- sixteen, the same the terminal theme from this colourway uses,
-- so the two look alike
local term = {
`
	terminalTail = `}
for i, v in ipairs(term) do vim.g["terminal_color_" .. (i - 1)] = v end
`
)

// Write puts the scheme in neovim's lua after a header, each line of
// which becomes a comment. Each text colour's comment carries its
// contrast with the ground, worked out here, so the number beside it is
// never stale.
func (scheme Scheme) Write(out io.Writer, header string) error {
	var text strings.Builder
	for line := range strings.SplitSeq(strings.TrimSpace(header), "\n") {
		fmt.Fprintln(&text, strings.TrimRight("-- "+line, " "))
	}
	fmt.Fprintf(&text, "--\n--   :colorscheme %s\n\n", scheme.Name)
	text.WriteString(preamble)
	background := "dark"
	if scheme.Light {
		background = "light"
	}
	fmt.Fprintf(&text, "vim.o.background = %q\n", background)
	fmt.Fprintf(&text, "vim.g.colors_name = %q\n\n", scheme.Name)
	fmt.Fprintln(&text, "local c = {")
	ground := scheme.Colours["bg"]
	for _, entry := range Roles {
		value := scheme.Colours[entry.Name]
		note := entry.Use
		if entry.Name != "bg" {
			ratio := Contrast(value, ground)
			note = fmt.Sprintf("%5.2f:1  %s", ratio, entry.Use)
		}
		fmt.Fprintf(&text, "  %-8s = %q, -- %s\n", entry.Name,
			ghostty.Hex(value), note)
	}
	fmt.Fprintln(&text, "}")
	fmt.Fprintln(&text)
	text.WriteString(groups)
	fmt.Fprintln(&text)
	text.WriteString(terminalHead)
	for row := 0; row < 16; row += 8 {
		quoted := make([]string, 8)
		for column := range 8 {
			value := scheme.Terminal[row+column]
			quoted[column] = strconv.Quote(ghostty.Hex(value))
		}
		fmt.Fprintf(&text, "  %s,\n", strings.Join(quoted, ", "))
	}
	text.WriteString(terminalTail)
	_, err := io.WriteString(out, text.String())
	return err
}

// Inverted is the scheme with every colour turned to its opposite, each
// lamp 255 less itself, for a screen the system inverts.
func (scheme Scheme) Inverted() Scheme {
	inverted := Scheme{Name: scheme.Name,
		Colours: map[string]swatch.Swatch{}}
	for _, value := range scheme.Terminal {
		inverted.Terminal = append(inverted.Terminal, invert(value))
	}
	for name, value := range scheme.Colours {
		inverted.Colours[name] = invert(value)
	}
	inverted.Light = lighter(inverted.Colours)
	return inverted
}

// Contrast is wcag's ratio between two colours: their relative
// luminances, each with 0.05 added, the lighter over the darker.
func Contrast(first, second swatch.Swatch) float64 {
	lighter, darker := luminance(first), luminance(second)
	if lighter < darker {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05)
}

// luminance is a colour's relative luminance: its y, against white's.
func luminance(value swatch.Swatch) float64 {
	_, y, _ := value.XYZ()
	_, white, _ := srgb.MustHex("#ffffff").Swatch().XYZ()
	return y / white
}

// parse reads a #rrggbb.
func parse(hex string) (swatch.Swatch, error) {
	lamps, err := srgb.FromHex(hex)
	if err != nil {
		return swatch.Swatch{}, err
	}
	return lamps.Swatch(), nil
}

// invert is a colour's opposite on the lamps.
func invert(value swatch.Swatch) swatch.Swatch {
	lamps, _ := srgb.FromSwatch(value)
	r, g, b := lamps.Bytes()
	return srgb.RGB8(255-r, 255-g, 255-b).Swatch()
}
