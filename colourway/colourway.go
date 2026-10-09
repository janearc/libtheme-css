// Package colourway is a colourway's source, one css sheet, and the
// files it is rendered into through the dialects: ghostty's theme,
// neovim's and vim's schemes and glamour's style, each with an inverted
// twin.
//
// The source is the truth. A rendered file says which source it came
// from and is never edited; a tool that changes a colourway changes its
// source, writes it back with Write, and renders again.
package colourway

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/janearc/libtheme-css/css"
)

// Source is a colourway's source: its name, the words it opens with, its
// roles, and the path it is named by in what it renders.
type Source struct {
	Name   string
	About  string
	Roles  *css.Sheet
	Origin string
}

// Read reads a source sheet. Its name is the file's, without .css.
func Read(path string) (Source, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Source{}, err
	}
	base := filepath.Base(path)
	return Source{Name: strings.TrimSuffix(base, filepath.Ext(base)),
		About: Opening(string(raw)), Roles: css.Read(string(raw)).Roles,
		Origin: "sources/" + base}, nil
}

// Write puts a source back as a sheet: its opening words, then its
// roles, each with its oklch beside it.
func (source Source) Write(path string) error {
	text := Commented(source.About) + source.Roles.String()
	return os.WriteFile(path, []byte(text), 0o644)
}

// Opening is a css file's first comment, as plain lines: the words a
// colourway opens with.
func Opening(text string) string {
	text = strings.TrimSpace(text)
	end := strings.Index(text, "*/")
	if !strings.HasPrefix(text, "/*") || end < 0 {
		return ""
	}
	lines := []string{}
	for line := range strings.SplitSeq(text[2:end], "\n") {
		line = strings.TrimPrefix(strings.TrimSpace(line), "*")
		lines = append(lines, strings.TrimSpace(line))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// Commented is words as a css comment, or nothing.
func Commented(about string) string {
	if about == "" {
		return ""
	}
	lines := strings.Split(about, "\n")
	for position := range lines {
		lines[position] = strings.TrimRight(" * "+lines[position], " ")
	}
	return "/*\n" + strings.Join(lines, "\n") + "\n */\n"
}
