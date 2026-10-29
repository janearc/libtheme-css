// Package colourway holds a colourway's source and the files rendered from
// it. The source is one css sheet. The rendered files are ghostty's theme,
// neovim's and vim's schemes and glamour's style, each with an inverted
// twin.
//
// The source is the truth. Every rendered file says which source it came
// from, and nobody edits it. To change a colourway, change its source, write
// it back with Write, and render again.
//
// The source holds sRGB. Its colours are the bytes in its hex; the oklch
// beside each only describes it. A source written from a colour between two
// bytes, or past what a screen can make, holds the nearest byte colour, and
// every file rendered from it says that colour.
package colourway

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/internal/plain"
)

// Source is a colourway's source. Name comes from the file name. About is
// the comment the file opens with. Roles are its colours. Origin is the
// path that rendered files name as their source.
type Source struct {
	Name   string
	About  string
	Roles  *css.Sheet
	Origin string
}

// nothing is the count of colours that could not be read that there should
// be. It is defined here.
const nothing = 0

// Read reads a source sheet. Its name is the file's without .css. Colours
// in it that are not complete ones are errors, since the role would
// silently be missing from every file rendered. So are roles written in
// another form of colour, such as oklch or rgb.
func Read(path string) (Source, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Source{}, err
	}
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if err := plain.Name(name); err != nil {
		return Source{}, fmt.Errorf("colourway: %s cannot name one: %w",
			strconv.Quote(base), err)
	}
	read := css.Read(string(raw))
	named := read.Named(isRole)
	refused := append(read.Skipped(), named...)
	if len(refused) > nothing {
		return Source{}, fmt.Errorf("colourway: %s: %s",
			strconv.Quote(base), strings.Join(refused, "; "))
	}
	return Source{Name: name, About: Opening(string(raw)),
		Roles: read.Roles, Origin: "sources/" + base}, nil
}

// Write saves the source to path as a css sheet. It writes the opening
// words as a comment, then the roles, each with its oklch beside it.
func (source Source) Write(path string) error {
	text := Commented(source.About) + source.Roles.String()
	return os.WriteFile(path, []byte(text), 0o644)
}

// Opening returns the text of a css file's first comment, without the
// comment marks. These are the words a colourway opens with. It returns an
// empty string if the file does not open with a comment.
func Opening(text string) string {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/*") {
		return ""
	}
	// The comment closes at the first "*/" after its own opening, so "/*/"
	// opens one and does not close it.
	words, _, closed := strings.Cut(text[len("/*"):], "*/")
	if !closed {
		return ""
	}
	lines := []string{}
	for line := range strings.SplitSeq(words, "\n") {
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
	lines := strings.Split(strings.ReplaceAll(about, "*/", "* /"), "\n")
	for position := range lines {
		lines[position] = strings.TrimRight(" * "+lines[position], " ")
	}
	return "/*\n" + strings.Join(lines, "\n") + "\n */\n"
}
