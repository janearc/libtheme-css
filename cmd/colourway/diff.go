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

// rolesOf returns the roles of any file colourway reads. The name picks the
// reader: .css is a source, .lua a neovim scheme, .json a markdown style.
// Anything else is a terminal theme. /dev/null gives no roles, because git
// passes it for a file that was added or removed.
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

// change is one role compared. It holds the name, the colour before and
// after, and whether each side has the role.
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

// report writes how two colourways differ. It prints a count, then a line
// for each role that changed, was added or was removed, with its hex before
// and after. If colour is true, it draws a swatch beside each. It returns
// true if anything differed.
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

// changedLine writes one changed role, before and after. It shows how far
// the role moved in oklab, times a hundred. It adds a note when the move is
// under ok.Eye, the smallest difference an eye can tell. It also shows the
// change in lightness, chroma and hue.
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

// swatchOf returns two cells lit in the colour, then a space. It returns
// an empty string if colour is false.
func swatchOf(value swatch.Swatch, colour bool) string {
	if !colour {
		return ""
	}
	var r, g, b int
	fmt.Sscanf(hexOf(value), "#%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm  \x1b[0m ", r, g, b)
}

// diff is colourway diff. It compares two files, or git's arguments with
// --git. Outside git it returns true if they differed, and main exits 1 on
// that, as diff does. Inside git it always succeeds, because git takes a
// failure to mean the diff broke. It prints any trouble instead.
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

// gitDiff runs the diff as git's external driver asks for it. Git passes
// seven arguments for a changed file: path, old file, old hex, old mode,
// new file, new hex, new mode. It passes nine for a renamed file and one
// for an unmerged path.
//
// It prints trouble instead of returning it, so git goes on.
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

// terminal returns true if stdout is a terminal. A swatch is then drawn
// without being asked for.
func terminal() bool {
	stat, err := os.Stdout.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice != 0
}

// show lists each role a file has, with hex and oklch values.
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
