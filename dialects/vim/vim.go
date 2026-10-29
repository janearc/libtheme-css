// Package vim is the vim dialect: a colourway's scheme said in vimscript,
// for vim, which does not read neovim's lua. It is neovim's scheme, with
// the same colours on the same groups. It leaves out the groups only
// neovim has, which are the tree-sitter names. Vim cannot hold them.
package vim

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/internal/plain"
)

// note says what vim needs to draw a scheme's colours in a terminal.
const note = `" in a terminal, vim draws these colours ` +
	`when termguicolors is set,
" in your vimrc:
"
"   set termguicolors

`

// vimName is a group name vim can hold: a letter, then letters, digits
// and underscores.
var vimName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// Write puts a scheme in vimscript after a header. Each line of the header
// becomes a comment.
func Write(out io.Writer, scheme nvim.Scheme, header string) error {
	if err := plain.Name(scheme.Name); err != nil {
		return fmt.Errorf("vim: scheme name: %w", err)
	}
	if err := scheme.Complete(); err != nil {
		return err
	}
	var text strings.Builder
	text.WriteString(plain.Comment(`"`, header))
	fmt.Fprintf(&text, "\"\n\"   :colorscheme %s\n\"\n", scheme.Name)
	text.WriteString(note)
	background := "dark"
	if scheme.Light {
		background = "light"
	}
	fmt.Fprintf(&text, "hi clear\nif exists('syntax_on')\n"+
		"  syntax reset\nendif\n")
	fmt.Fprintf(&text, "set background=%s\nlet g:colors_name = '%s'\n\n",
		background, strings.ReplaceAll(scheme.Name, "'", "''"))
	for _, group := range nvim.Groups() {
		if line, has := highlight(group, scheme); has {
			fmt.Fprintln(&text, line)
		}
	}
	text.WriteString(terminal(scheme))
	_, err := io.WriteString(out, text.String())
	return err
}

// highlight is one group as vim's highlight command. The bool says
// whether vim can hold the group at all.
func highlight(group nvim.Group, scheme nvim.Scheme) (string, bool) {
	if !vimName.MatchString(group.Name) {
		return "", false
	}
	parts := []string{"hi", group.Name}
	if value, has := scheme.Colours[group.Fg]; group.Fg != "" && has {
		parts = append(parts, "guifg="+ghostty.Hex(value))
	}
	if value, has := scheme.Colours[group.Bg]; group.Bg != "" && has {
		parts = append(parts, "guibg="+ghostty.Hex(value))
	}
	styles := "NONE"
	if len(group.Styles) > 0 {
		styles = strings.Join(group.Styles, ",")
	}
	parts = append(parts, "gui="+styles, "cterm="+styles)
	return strings.Join(parts, " "), true
}

// terminal is the sixteen for vim's own terminal, so a shell inside vim
// (:terminal) draws like the terminal around it. The scheme is complete, so
// it has all sixteen.
func terminal(scheme nvim.Scheme) string {
	quoted := make([]string, len(scheme.Terminal))
	for number, value := range scheme.Terminal {
		quoted[number] = "'" + ghostty.Hex(value) + "'"
	}
	return "\nlet g:terminal_ansi_colors = [\n  \\ " +
		strings.Join(quoted[:8], ", ") + ",\n  \\ " +
		strings.Join(quoted[8:], ", ") + "]\n"
}
