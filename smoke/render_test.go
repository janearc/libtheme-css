package smoke_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/dialects/vim"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// distinct are sixteen colours that differ from each other, one for each
// role of a scheme, so a group drawn in the wrong role shows.
var distinct = []string{"#101820", "#203040", "#305060", "#407080", "#5090a0",
	"#60b0c0", "#70d0e0", "#80f0ff", "#a01820", "#b03040", "#c05060",
	"#d07080", "#e09090", "#f0b0a0", "#e0d0b0", "#d0f0c0"}

// coloured is a scheme with a different colour for every role.
func coloured(t *testing.T) nvim.Scheme {
	t.Helper()
	roles := css.New()
	for position, entry := range nvim.Roles {
		roles.Set(entry.Role, srgb.MustHex(distinct[position]).Swatch())
	}
	for _, each := range ghostty.Ansi {
		roles.Set(each, srgb.MustHex(grey).Swatch())
	}
	scheme, err := nvim.Of("coloured", roles)
	if err != nil {
		t.Fatal(err)
	}
	return scheme
}

// hexOf is a swatch as the files write it.
func hexOf(value swatch.Swatch) string { return ghostty.Hex(value) }

// drawn is what the model says a group draws with: text and ground, or
// empty for a colour the group leaves to the editor.
type drawn struct {
	text, ground string
}

// expected is every group the model reads, with the colours it gives it.
func expected(scheme nvim.Scheme) map[string]drawn {
	groups := map[string]drawn{}
	for _, group := range nvim.Groups() {
		var one drawn
		if group.Fg != "" {
			one.text = hexOf(scheme.Colours[group.Fg])
		}
		if group.Bg != "" {
			one.ground = hexOf(scheme.Colours[group.Bg])
		}
		groups[group.Name] = one
	}
	return groups
}

// dumpLua writes each group's text and ground, as neovim holds them, to the
// report: a line of the name, the text and the ground, with - for none.
const dumpLua = `local names = {}
for name in pairs(vim.api.nvim_get_hl(0, {})) do names[#names + 1] = name end
table.sort(names)
local out = io.open(vim.env.REPORT, "w")
for _, name in ipairs(names) do
  local hl = vim.api.nvim_get_hl(0, {name = name, link = false})
  local fg = hl.fg and string.format("#%06x", hl.fg) or "-"
  local bg = hl.bg and string.format("#%06x", hl.bg) or "-"
  out:write(name, " ", fg, " ", bg, "\n")
end
out:close()
`

// dumpVim is the same for vim, for the names in the list the caller sets.
const dumpVim = `let out = []
for name in g:names
  let id = synIDtrans(hlID(name))
  let text = synIDattr(id, 'fg#', 'gui')
  let ground = synIDattr(id, 'bg#', 'gui')
  call add(out, name . ' ' . (empty(text) ? '-' : text) . ' ' . (empty(ground) ? '-' : ground))
endfor
call writefile(out, $REPORT)
`

// held reads a dump into what each group draws. A dash is a colour left to
// the editor, and so is a name such as Red: vim reports its own terminal
// default that way for a group with no gui colour.
func held(t *testing.T, report string) map[string]drawn {
	t.Helper()
	groups := map[string]drawn{}
	for _, line := range strings.Split(strings.TrimSpace(report), "\n") {
		fields := strings.Fields(line)
		one := drawn{}
		one.text, one.ground = colourAt(fields, textAt), colourAt(fields, groundAt)
		groups[fields[nameAt]] = one
	}
	return groups
}

// textAt and groundAt are the places in a dump line of the text colour and the
// ground colour. They are defined by the dump scripts above.
const (
	// nameAt is where the group's name is. It is defined by the dump scripts.
	nameAt = 0
	// textAt is where the text colour is. It is defined by the dump scripts.
	textAt = 1
	// groundAt is where the ground colour is. It is defined by the dump
	// scripts.
	groundAt = 2
)

// colourAt is the hex colour in a place of a dump line, or empty if the
// place has none.
func colourAt(fields []string, place int) string {
	if place < len(fields) && strings.HasPrefix(fields[place], "#") {
		return fields[place]
	}
	return ""
}

// dumped runs the editor over the scheme file and the dump file, with the
// report's path in the environment, and returns what the report holds.
func dumped(t *testing.T, editor string, arguments []string, files ...string) string {
	t.Helper()
	found, err := exec.LookPath(editor)
	if err != nil {
		t.Fatalf("this test needs %s on the path: %v", editor, err)
	}
	report := filepath.Join(t.TempDir(), "report.txt")
	run := exec.Command(found, arguments...)
	run.Env = append(os.Environ(), "REPORT="+report)
	if output, err := run.CombinedOutput(); err != nil {
		t.Fatalf("%s failed: %v\n%s", editor, err, output)
	}
	raw, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// scripted writes text to a file of its own and returns the path.
func scripted(t *testing.T, name, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(text), private); err != nil {
		t.Fatal(err)
	}
	return path
}

// The smoke test: every group the model reads is drawn by a real neovim in
// the colours the model gives it, and the groups the model says leave a
// colour to the editor leave it. The acceptance gate holds the model's
// pairs, so this is where the model is held to the editor.
func TestNeovimDrawsEveryGroupAsTheModelSays(t *testing.T) {
	scheme := coloured(t)
	var lua strings.Builder
	if err := scheme.Write(&lua, "header"); err != nil {
		t.Fatal(err)
	}
	got := inNeovim(t, lua.String())
	for name, want := range expected(scheme) {
		if got[name] != want {
			t.Errorf("neovim draws %s as %+v, the model says %+v", name, got[name], want)
		}
	}
}

// vimName is a group name vim can hold, as the vim dialect keeps it.
var vimName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// The same for vim, for the groups vim can hold.
func TestVimDrawsEveryGroupAsTheModelSays(t *testing.T) {
	scheme := coloured(t)
	var script strings.Builder
	if err := vim.Write(&script, scheme, "header"); err != nil {
		t.Fatal(err)
	}
	var names []string
	for name := range expected(scheme) {
		if vimName.MatchString(name) {
			names = append(names, "'"+name+"'")
		}
	}
	setNames := fmt.Sprintf("let g:names = [%s]\n", strings.Join(names, ", "))
	got := inVimDump(t, script.String(), setNames+dumpVim)
	for name, want := range expected(scheme) {
		if vimName.MatchString(name) && got[name] != want {
			t.Errorf("vim draws %s as %+v, the model says %+v", name, got[name], want)
		}
	}
}

// The anti-test: the dump sees a wrong colour. A scheme with one group
// drawn in another role would differ from the model, so the model is not
// compared with itself.
func TestTheDumpSeesAColourThatIsWrong(t *testing.T) {
	scheme := coloured(t)
	other := coloured(t)
	other.Colours["ink"] = other.Colours["err"]
	var lua strings.Builder
	if err := other.Write(&lua, "header"); err != nil {
		t.Fatal(err)
	}
	got := inNeovim(t, lua.String())
	if got["Normal"] == expected(scheme)["Normal"] {
		t.Error("a scheme with another ink drew Normal as the model says for the first")
	}
}

// inNeovim loads the scheme in neovim and returns what it draws each group
// with.
func inNeovim(t *testing.T, scheme string) map[string]drawn {
	t.Helper()
	arguments := []string{"--headless", "-u", "NONE", "-i", "NONE",
		"-c", "luafile " + scripted(t, "scheme.lua", scheme),
		"-c", "luafile " + scripted(t, "dump.lua", dumpLua), "-c", "qa!"}
	return held(t, dumped(t, "nvim", arguments))
}

// inVimDump loads the scheme in vim, runs the dump script, and returns what
// it draws each group with.
func inVimDump(t *testing.T, scheme, dump string) map[string]drawn {
	t.Helper()
	arguments := []string{"-es", "-u", "NONE", "-i", "NONE", "-N",
		"-c", "source " + scripted(t, "scheme.vim", scheme),
		"-c", "source " + scripted(t, "dump.vim", dump), "-c", "qa!"}
	return held(t, dumped(t, "vim", arguments))
}
