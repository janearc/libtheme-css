package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/glamour"
	"github.com/janearc/libtheme-css/dialects/nvim"
)

// sheet makes a source from programs' files and prints it. The order is the
// terminal's settings and palette, claude code's colours named from that
// palette, the editor's table, then the markdown style's colours. If a later
// file sets a role differently, it reports that and keeps the first.
func sheet(arguments []string, to streams) error {
	files, err := sheetFiles(arguments)
	if err != nil {
		return err
	}
	merged, about := css.New(), ""
	if path, found := files["--ghostty"]; found {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		theme, err := ghostty.Read(bytes.NewReader(raw))
		if err != nil {
			return err
		}
		merge(to.errs, merged, terminalRoles(theme), path, "")
		about = comments(string(raw))
	}
	if path, found := files["--nvim"]; found {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scheme, err := nvim.Read(file)
		if err != nil {
			return err
		}
		merge(to.errs, merged, scheme.Roles(), path, "nvim-")
	}
	if path, found := files["--glamour"]; found {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		roles, err := glamour.Read(raw)
		if err != nil {
			return err
		}
		merge(to.errs, merged, roles, path, "")
	}
	fmt.Fprint(to.out, colourway.Commented(about)+merged.String())
	return nil
}

// sheetOptions are the programs sheet can read a file of.
var sheetOptions = []string{"--ghostty", "--nvim", "--glamour"}

// pair is an option and its file. It is defined by the usage.
const pair = 2

// sheetFiles is the file for each option, from the arguments. An option
// nobody knows, one without a file, and no option at all are wrong calls.
func sheetFiles(arguments []string) (map[string]string, error) {
	files := map[string]string{}
	for ; len(arguments) >= pair; arguments = arguments[pair:] {
		option := arguments[0]
		if !slices.Contains(sheetOptions, option) {
			return nil, usageError{"sheet has no option " + option}
		}
		files[option] = arguments[1]
	}
	if len(arguments) != nothing || len(files) == nothing {
		return nil, usageError{"sheet wants an option and its file"}
	}
	return files, nil
}

// terminalRoles returns a theme's roles with claude code's colours named.
// It leaves out their palette numbers, because the names say them.
func terminalRoles(theme ghostty.Theme) *css.Sheet {
	named := claude.Of(theme.Palette)
	roles := css.New()
	for _, rule := range theme.Roles().Rules() {
		if !claudeNumber(rule.Name) {
			roles.Set(rule.Name, rule.Swatch)
		}
	}
	for _, rule := range named.Rules() {
		roles.Set(rule.Name, rule.Swatch)
	}
	return roles
}

// claudeNumber is whether a palette-N role is one claude code draws in.
func claudeNumber(name string) bool {
	for _, slot := range claude.Slots {
		number := fmt.Sprintf("palette-%d", slot.Number)
		if slot.Number >= 16 && name == number {
			return true
		}
	}
	return false
}

// merge copies every role of one sheet into another. If both have a role
// with different values, the first is kept. The second is also kept under
// the program's own prefix, such as nvim-ink, when the program has one.
// Either way it prints a note to errs.
func merge(errs io.Writer, into, from *css.Sheet, file, own string) {
	for _, rule := range from.Rules() {
		held, found := into.Get(rule.Name)
		switch {
		case !found:
			into.Set(rule.Name, rule.Swatch)
		case ghostty.Hex(held) == rule.Value:
		case own != "":
			into.Set(own+rule.Name, rule.Swatch)
			fmt.Fprintf(errs,
				"colourway: %s: %s says %s; kept as %s\n",
				rule.Name, file, rule.Value, own+rule.Name)
		default:
			fmt.Fprintf(errs,
				"colourway: %s: %s says %s; kept %s\n",
				rule.Name, file, rule.Value, ghostty.Hex(held))
		}
	}
}

// comments returns a theme file's opening comment lines, without their
// marks.
func comments(text string) string {
	lines := []string{}
	for line := range strings.SplitSeq(text, "\n") {
		if !strings.HasPrefix(line, "#") {
			break
		}
		line = strings.TrimPrefix(line, "#")
		lines = append(lines, strings.TrimSpace(line))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
