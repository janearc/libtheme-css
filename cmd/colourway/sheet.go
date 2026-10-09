package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/glamour"
	"github.com/janearc/libtheme-css/dialects/nvim"
)

// sheet makes a source from programs' files and prints it: the
// terminal's settings and palette, claude code's colours named out of
// that palette, the editor's table, then the markdown style's colours.
// A role a later file sets differently is reported, and the first kept.
func sheet(arguments []string) error {
	files := map[string]string{}
	for ; len(arguments) >= 2; arguments = arguments[2:] {
		files[arguments[0]] = arguments[1]
	}
	if len(arguments) != 0 || len(files) == 0 {
		usage()
		os.Exit(2)
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
		merge(merged, terminalRoles(theme), path, "")
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
		merge(merged, scheme.Roles(), path, "nvim-")
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
		merge(merged, roles, path, "")
	}
	fmt.Print(colourway.Commented(about) + merged.String())
	return nil
}

// terminalRoles are a theme's roles with claude code's colours named,
// and their palette numbers left out, since the names say them.
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

// merge sets every role of one sheet into another. Where the two
// disagree, the first is kept, and the second is kept beside it under
// its program's own name, nvim-ink, when the program has one; either
// way it is said.
func merge(into, from *css.Sheet, file, own string) {
	for _, rule := range from.Rules() {
		held, found := into.Get(rule.Name)
		switch {
		case !found:
			into.Set(rule.Name, rule.Swatch)
		case ghostty.Hex(held) == rule.Value:
		case own != "":
			into.Set(own+rule.Name, rule.Swatch)
			fmt.Fprintf(os.Stderr,
				"colourway: %s: %s says %s; kept as %s\n",
				rule.Name, file, rule.Value, own+rule.Name)
		default:
			fmt.Fprintf(os.Stderr,
				"colourway: %s: %s says %s; kept %s\n",
				rule.Name, file, rule.Value, ghostty.Hex(held))
		}
	}
}

// comments are a theme file's opening comment lines, without their
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
