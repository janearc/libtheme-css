package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/glamour"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
)

// rolesOf is a colourway's roles from any file colourway reads, told by
// its name: a source (.css), a vim scheme (.lua), a markdown style
// (.json), or else a terminal theme. /dev/null, which git gives for a
// file added or removed, is no roles.
func rolesOf(path string) (*css.Sheet, error) {
	if path == os.DevNull {
		return css.New(), nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	switch filepath.Ext(path) {
	case ".css":
		source, err := colourway.Read(path)
		return source.Roles, err
	case ".lua":
		scheme, err := nvim.Read(bytes.NewReader(raw))
		return scheme.Roles(), err
	case ".json":
		return glamour.Read(raw)
	}
	theme, err := ghostty.Read(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return terminalRoles(theme), nil
}

// change is one role compared: its name, its colour before and after, and
// which of the two it has.
type change struct {
	name          string
	before, after swatch.Swatch
	had, has      bool
}

// compared is every role either sheet has, the first sheet's order first.
func compared(before, after *css.Sheet) []change {
	changes, seen := []change{}, map[string]bool{}
	for _, sheet := range []*css.Sheet{before, after} {
		for _, name := range sheet.Names() {
			if seen[name] {
				continue
			}
			seen[name] = true
			was, had := before.Get(name)
			is, has := after.Get(name)
			changes = append(changes,
				change{name, was, is, had, has})
		}
	}
	return changes
}

// report writes how two colourways differ: a count, then each role that
// changed, came or went, with its hex before and after. Colour draws a
// swatch beside each. It says whether anything differed.
func report(out io.Writer, changes []change, colour bool) bool {
	same, changed, added, removed := 0, 0, 0, 0
	lines := []string{}
	for _, one := range changes {
		both := one.had && one.has
		switch {
		case both && hexOf(one.before) == hexOf(one.after):
			same++
		case both:
			changed++
			lines = append(lines, changedLine(one, colour))
		case one.has:
			added++
			lines = append(lines,
				single("+", one.name, one.after, colour))
		default:
			removed++
			lines = append(lines,
				single("-", one.name, one.before, colour))
		}
	}
	fmt.Fprintf(out, "%d the same, %d changed, %d added, %d removed\n",
		same, changed, added, removed)
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}
	return changed+added+removed > 0
}

// single is a role that came or went, and its colour.
func single(mark, name string, value swatch.Swatch, colour bool) string {
	return fmt.Sprintf("%s %-18s %s%s", mark, name, swatchOf(value, colour),
		hexOf(value))
}

// changedLine is one changed role, before and after: how far it moved in
// oklab, times a hundred, with a note when an eye could not tell (ok.Eye
// is the space author's just-noticeable difference), and the change in
// lightness, chroma and hue.
func changedLine(one change, colour bool) string {
	was, is := ok.FromSwatch(one.before), ok.FromSwatch(one.after)
	distance := ok.Distance(was, is)
	note := ""
	if distance < ok.Eye {
		note = "  (under what an eye can tell)"
	}
	before, after := was.Polar(), is.Polar()
	hue := math.Mod(after.H-before.H+540, 360) - 180
	shape := "~ %-18s %s%s -> %s%s  moved %.1f  " +
		"L %+.2f C %+.3f H %+.0f%s"
	return fmt.Sprintf(shape, one.name,
		swatchOf(one.before, colour), hexOf(one.before),
		swatchOf(one.after, colour), hexOf(one.after), distance*100,
		after.L-before.L, after.C-before.C, hue, note)
}

// hexOf is a colour as the files write it.
func hexOf(value swatch.Swatch) string {
	return strings.ToLower(ghostty.Hex(value))
}

// swatchOf is two cells lit in a colour, then a space, when colour is
// wanted, and nothing otherwise.
func swatchOf(value swatch.Swatch, colour bool) string {
	if !colour {
		return ""
	}
	var r, g, b int
	fmt.Sscanf(hexOf(value), "#%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm  \x1b[0m ", r, g, b)
}

// diff is colourway diff: two files, or git's arguments with --git,
// compared. Outside git it says whether they differed by its exit, as
// diff does. Inside git it always succeeds, since git takes a failure as
// the diff breaking, and says any trouble in the diff instead.
func diff(arguments []string) (bool, error) {
	colour, git := terminal(), false
	for len(arguments) > 0 && strings.HasPrefix(arguments[0], "--") {
		colour = colour || arguments[0] == "--colour"
		git = git || arguments[0] == "--git"
		arguments = arguments[1:]
	}
	if git {
		gitDiff(arguments, colour)
		return false, nil
	}
	if len(arguments) != 2 {
		usage()
		os.Exit(2)
	}
	return compare(arguments[0], arguments[1], colour)
}

// gitDiff is the diff as git's external driver asks for it: seven
// arguments for a file changed (path, then the old file, hex and mode,
// then the new), nine for one renamed, one for a path unmerged.
//
// Trouble is written, not returned, so git goes on.
func gitDiff(arguments []string, colour bool) {
	switch len(arguments) {
	case 1:
		fmt.Printf("colourway diff %s: unmerged\n", arguments[0])
		return
	case 7:
		fmt.Printf("colourway diff %s\n", arguments[0])
	case 9:
		fmt.Printf("colourway diff %s -> %s\n",
			arguments[0], arguments[7])
	default:
		fmt.Printf("colourway diff: git gave %d arguments\n",
			len(arguments))
		return
	}
	_, err := compare(arguments[1], arguments[4], colour)
	if err != nil {
		fmt.Printf("colourway: %v\n", err)
	}
}

// compare reads two colourways and reports how they differ.
func compare(beforePath, afterPath string, colour bool) (bool, error) {
	before, err := rolesOf(beforePath)
	if err != nil {
		return false, err
	}
	after, err := rolesOf(afterPath)
	if err != nil {
		return false, err
	}
	return report(os.Stdout, compared(before, after), colour), nil
}

// terminal is whether what is written goes straight to a terminal, which
// is when a swatch can be drawn without being asked for.
func terminal() bool {
	stat, err := os.Stdout.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice != 0
}

// show is colourway show: one line for each role a file has, its hex and
// its oklch, which is the form git compares colourways in.
func show(path string) error {
	roles, err := rolesOf(path)
	if err != nil {
		return err
	}
	for _, rule := range roles.Rules() {
		polar := ok.FromSwatch(rule.Swatch).Polar()
		fmt.Printf("%-18s %s  oklch(%.0f%% %.3f %.1f)\n", rule.Name,
			hexOf(rule.Swatch), polar.L*100, polar.C, polar.H)
	}
	return nil
}
