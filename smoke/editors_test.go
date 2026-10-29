package smoke_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// private is the permission a scratch file is written with: the owner's
// read and write, nobody else's. It is defined here.
const private = os.FileMode(0o600)

// left is what a loaded scheme left behind: whether the variable called
// injected exists, which an honest file never sets, and the colour
// scheme's name as the editor holds it.
type left struct {
	injected bool
	name     string
}

// vimProbe writes what vim holds to a report file.
const vimProbe = "call writefile([exists('g:injected') ? 'yes' : 'no', " +
	"get(g:, 'colors_name', 'none')], '%s')"

// nvimProbe writes what neovim holds to a report file.
const nvimProbe = `lua local out = io.open("%s", "w"); ` +
	`out:write(rawget(_G, "injected") ~= nil and "yes" or "no", "\n", ` +
	`vim.g.colors_name or "none", "\n"); out:close()`

// inVim sources a vimscript file in vim.
func inVim(t *testing.T, script string) left {
	t.Helper()
	return load(t, "vim", "scheme.vim", script, vimPlan)
}

// vimPlan is the arguments that start vim on a script and have it report.
func vimPlan(file, report string) []string {
	return []string{"-es", "-u", "NONE", "-i", "NONE", "-N", "-c", "source " + file,
		"-c", fmt.Sprintf(vimProbe, report), "-c", "qa!"}
}

// inNvim runs a lua file in neovim.
func inNvim(t *testing.T, script string) left {
	t.Helper()
	return load(t, "nvim", "scheme.lua", script, nvimPlan)
}

// nvimPlan is the arguments that start neovim on a script and have it
// report.
func nvimPlan(file, report string) []string {
	return []string{"--headless", "-u", "NONE", "-i", "NONE",
		"-c", "luafile " + file, "-c", fmt.Sprintf(nvimProbe, report), "-c", "qa!"}
}

// load writes the script to a file, runs the editor with the arguments the
// plan makes from the script's path and the report's, and reads the report.
func load(t *testing.T, editor, name, script string,
	plan func(file, report string) []string) left {
	t.Helper()
	found, err := exec.LookPath(editor)
	if err != nil {
		t.Fatalf("this test needs %s on the path: %v", editor, err)
	}
	dir := t.TempDir()
	file, report := filepath.Join(dir, name), filepath.Join(dir, "report.txt")
	if err := os.WriteFile(file, []byte(script), private); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(found, plan(file, report)...).CombinedOutput(); err != nil {
		t.Fatalf("%s failed: %v\n%s", editor, err, output)
	}
	raw, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	injected, held, _ := strings.Cut(string(raw), "\n")
	return left{injected: injected == "yes", name: strings.TrimSuffix(held, "\n")}
}
