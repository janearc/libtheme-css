// Package colourway holds a colourway's source and the files rendered from
// it. The source is one css sheet. The rendered files are ghostty's theme,
// neovim's and vim's schemes and glamour's style, each with an inverted
// twin.
//
// The source is the truth. Every rendered file says which source it came
// from, and nobody edits it. To change a colourway, change its source, write
// it back with Write, and render again.
package colourway

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/janearc/libtheme-css/css"
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
