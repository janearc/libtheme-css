package main

import (
	"bytes"
	"flag"
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

// side is one side of a comparison: the name that says which dialect the
// bytes are in, and the path the bytes are read from. They are one path
// for a file given on the command line.
//
// Git hands the driver temp files whose names are not always the tracked
// file's, so it also gives the tracked name, and the tracked name wins.
type side struct {
	name, path string
}

// nothing is the count of colours, or of bytes, that there are none of. It
// is defined here.
const nothing = 0

// dialect is a kind of file colourway reads: what to call it, and how to
// read its bytes into roles.
type dialect struct {
	says string
	read func(from side, raw []byte) (*css.Sheet, error)
}

// dialects are the kinds of file by the extension of the name. A name with
// none of these is a terminal theme.
var dialects = map[string]dialect{
	".css":  {"a css source", readCSS},
	".lua":  {"a neovim scheme", readLua},
	".json": {"a markdown style", readMarkdown},
}

// terminalTheme is what a name with no extension of the above is.
var terminalTheme = dialect{"a terminal theme", readTerminal}

// dialectOf is the dialect a file name says its file is in.
func dialectOf(name string) dialect {
	key := strings.ToLower(filepath.Ext(name))
	if found, known := dialects[key]; known {
		return found
	}
	return terminalTheme
}

// readCSS reads a css source's roles.
func readCSS(from side, _ []byte) (*css.Sheet, error) {
	source, err := colourway.Read(from.path)
	return source.Roles, err
}

// readMarkdown reads a markdown style's roles.
func readMarkdown(_ side, raw []byte) (*css.Sheet, error) {
	return glamour.Read(raw)
}

// readLua reads a neovim scheme's roles.
func readLua(_ side, raw []byte) (*css.Sheet, error) {
	scheme, err := nvim.Read(bytes.NewReader(raw))
	return scheme.Roles(), err
}

// readTerminal reads a terminal theme's roles.
func readTerminal(_ side, raw []byte) (*css.Sheet, error) {
	theme, err := ghostty.Read(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return terminalRoles(theme), nil
}

// rolesOf returns the roles of any file colourway reads, in the dialect the
// side's name says. /dev/null gives no roles, because git passes it for a
// file that was added or removed.
//
// A file that has text in it and gives no roles was read as the wrong
// dialect, or is not a colourway. That is an error, not an empty file.
func rolesOf(from side) (*css.Sheet, error) {
	if from.path == os.DevNull {
		return css.New(), nil
	}
	raw, err := os.ReadFile(from.path)
	if err != nil {
		return nil, err
	}
	kind := dialectOf(from.name)
	roles, err := kind.read(from, raw)
	if err != nil {
		return nil, err
	}
	blank := len(bytes.TrimSpace(raw)) == nothing
	if len(roles.Names()) == nothing && !blank {
		return nil, fmt.Errorf("%s has text and no colours, read as %s",
			from.name, kind.says)
	}
	return roles, nil
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
	if unseen(distance) {
		note = "  (under what an eye can tell)"
	}
	before, after := was.Polar(), is.Polar()
	hue := hueMoved(before.H, after.H)
	shape := "~ %-18s %s%s -> %s%s  moved %.1f  " +
		"L %+.2f C %+.3f H %+.0f%s"
	return fmt.Sprintf(shape, one.name,
		swatchOf(one.before, colour), hexOf(one.before),
		swatchOf(one.after, colour), hexOf(one.after), distance*100,
		after.L-before.L, after.C-before.C, hue, note)
}

// unseen is whether a move in oklab is under what an eye can tell.
func unseen(distance float64) bool { return distance < ok.Eye }

// hueMoved is how far a hue went in degrees. It goes the short way round
// the wheel, so the answer is never past a half turn either way.
func hueMoved(before, after float64) float64 {
	return math.Mod(after-before+540, 360) - 180
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
// --git.
//
// Outside git it returns errDiffered if they differed, and the command
// exits 1 on that, as diff does. Inside git it always succeeds, because git
// takes a failure to mean the diff broke. It prints any trouble instead.
func diff(arguments []string, to streams) error {
	flags := flag.NewFlagSet("diff", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	colour := flags.Bool("colour", terminal(), "a swatch beside each role")
	git := flags.Bool("git", false, "arguments are git's external diff's")
	if err := flags.Parse(arguments); err != nil {
		return usageError{"diff: " + err.Error()}
	}
	files := flags.Args()
	if *git {
		gitDiff(to.out, files, *colour)
		return nil
	}
	if len(files) != filesToCompare {
		return usageError{"diff wants two files"}
	}
	before, after := filed(files[0]), filed(files[1])
	differed, err := compare(to.out, before, after, *colour)
	if err == nil && differed {
		return errDiffered
	}
	return err
}

const (
	// filesToCompare is how many files diff compares. It is defined by the
	// usage.
	filesToCompare = 2
	// filesToShow is how many files show shows. It is defined by the usage.
	filesToShow = 1
)

// filed is a side whose name is its path, as a file given on the command
// line is.
func filed(path string) side {
	return side{name: path, path: path}
}

// Where git puts each argument it gives an external diff, and how many it
// gives for each kind of call. All are defined by git's documentation of
// GIT_EXTERNAL_DIFF.
const (
	// pathAt is where the tracked path is. It is defined by git.
	pathAt = 0
	// oldFileAt is where the old file is. It is defined by git.
	oldFileAt = 1
	// newFileAt is where the new file is. It is defined by git.
	newFileAt = 4
	// renamedAt is where a rename's new path is. It is defined by git.
	renamedAt = 7

	// unmergedCall is how many an unmerged path has. It is defined by git.
	unmergedCall = 1
	// changedCall is how many a changed file has. It is defined by git.
	changedCall = 7
	// renamedCall is how many a renamed file has. It is defined by git.
	renamedCall = 9
)

// gitDiff runs the diff as git's external driver asks for it. Git passes
// seven arguments for a changed file: path, old file, old hex, old mode,
// new file, new hex, new mode. It passes nine for a renamed file and one
// for an unmerged path.
//
// The path is the tracked name, and it says each side's dialect; the old
// and new files are temp copies that are only read. For a rename the old
// side has the old path and the new side the new.
//
// It prints trouble instead of returning it, so git goes on.
func gitDiff(out io.Writer, arguments []string, colour bool) {
	var before, after side
	switch len(arguments) {
	case unmergedCall:
		tracked := arguments[pathAt]
		fmt.Fprintf(out, "colourway diff %s: unmerged\n", tracked)
		return
	case changedCall:
		fmt.Fprintf(out, "colourway diff %s\n", arguments[pathAt])
		before = side{arguments[pathAt], arguments[oldFileAt]}
		after = side{arguments[pathAt], arguments[newFileAt]}
	case renamedCall:
		fmt.Fprintf(out, "colourway diff %s -> %s\n",
			arguments[pathAt], arguments[renamedAt])
		before = side{arguments[pathAt], arguments[oldFileAt]}
		after = side{arguments[renamedAt], arguments[newFileAt]}
	default:
		fmt.Fprintf(out, "colourway diff: git gave %d arguments\n",
			len(arguments))
		return
	}
	if _, err := compare(out, before, after, colour); err != nil {
		fmt.Fprintf(out, "colourway: %v\n", err)
	}
}

// compare reads two colourways and reports how they differ.
func compare(out io.Writer, before, after side, colour bool) (bool, error) {
	was, err := rolesOf(before)
	if err != nil {
		return false, err
	}
	is, err := rolesOf(after)
	if err != nil {
		return false, err
	}
	return report(out, compared(was, is), colour), nil
}

// terminal returns true if stdout is a terminal. A swatch is then drawn
// without being asked for.
func terminal() bool {
	stat, err := os.Stdout.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice != 0
}

// show lists each role a file has, with hex and oklch values.
func show(arguments []string, to streams) error {
	if len(arguments) != filesToShow {
		return usageError{"show wants a file"}
	}
	roles, err := rolesOf(filed(arguments[0]))
	if err != nil {
		return err
	}
	for _, rule := range roles.Rules() {
		polar := ok.FromSwatch(rule.Swatch).Polar()
		fmt.Fprintf(to.out, "%-18s %s  oklch(%.0f%% %.3f %.1f)\n",
			rule.Name, hexOf(rule.Swatch),
			polar.L*100, polar.C, polar.H)
	}
	return nil
}
