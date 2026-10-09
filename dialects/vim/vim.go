// Package vim is the vim dialect: a colourway's scheme said in vimscript,
// for vim, which does not read neovim's lua. It is neovim's scheme, the
// same colours on the same groups, less the groups only neovim has: the
// tree-sitter names, which vim cannot hold.
package vim

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
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

// Write puts a scheme in vimscript after a header, each line of which
// becomes a comment.
func Write(out io.Writer, scheme nvim.Scheme, header string) error {
	var text strings.Builder
	for line := range strings.SplitSeq(strings.TrimSpace(header), "\n") {
		fmt.Fprintln(&text, strings.TrimRight(`" `+line, " "))
	}
	fmt.Fprintf(&text, "\"\n\"   :colorscheme %s\n\"\n", scheme.Name)
	text.WriteString(note)
	background := "dark"
	if scheme.Light {
		background = "light"
	}
	fmt.Fprintf(&text, "hi clear\nif exists('syntax_on')\n"+
		"  syntax reset\nendif\n")
	fmt.Fprintf(&text, "set background=%s\nlet g:colors_name = '%s'\n\n",
		background, scheme.Name)
	for _, group := range nvim.Groups() {
		if line, has := highlight(group, scheme); has {
			fmt.Fprintln(&text, line)
		}
	}
	text.WriteString(terminal(scheme))
	_, err := io.WriteString(out, text.String())
	return err
}

// highlight is one group as vim's highlight command, and whether vim can
// hold the group at all.
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
// (:terminal) draws like the terminal around it; nothing, for a scheme
// that has no sixteen.
func terminal(scheme nvim.Scheme) string {
	if len(scheme.Terminal) != 16 {
		return ""
	}
	quoted := make([]string, 16)
	for number, value := range scheme.Terminal {
		quoted[number] = "'" + ghostty.Hex(value) + "'"
	}
	return "\nlet g:terminal_ansi_colors = [\n  \\ " +
		strings.Join(quoted[:8], ", ") + ",\n  \\ " +
		strings.Join(quoted[8:], ", ") + "]\n"
}
