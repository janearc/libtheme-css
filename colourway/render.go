package colourway

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/glamour"
	"github.com/janearc/libtheme-css/dialects/nvim"
	"github.com/janearc/libtheme-css/dialects/vim"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// programs are the files Render writes, in order. Each has the program's
// name and a function that writes one colourway under a directory and
// returns the path it wrote.
var programs = []struct {
	name  string
	write func(colourway Source, dir string) (string, error)
}{
	{"ghostty", writeGhostty},
	{"nvim", writeNvim},
	{"vim", writeVim},
	{"glamour", writeGlamour},
}

// Written is one file Render made, or the reason it made none. It holds the
// program, the colourway's name, and the path or the error.
type Written struct {
	Program, Name, Path string
	Err                 error
}

// Inverted is the source's twin with every role turned to its opposite. It
// is for a screen whose colours the system inverts. Its name ends in
// -inverted.
func (source Source) Inverted() Source {
	return Source{Name: source.Name + "-inverted",
		About: source.About + "\n\ninverted, for a screen whose " +
			"colours the system inverts, so it shows as meant.",
		Roles: inverted(source.Roles), Origin: source.Origin}
}

// Render writes every file a source can fill, under dir, for the source and
// for its inverted twin. It returns what it wrote and what it skipped, with
// the reason.
//
// Each program renders from the resolved roles, so a general role the source
// does not set is derived. If a role cannot be derived, that file is skipped.
func Render(source Source, dir string) []Written {
	written := []Written{}
	for _, each := range []Source{source, source.Inverted()} {
		for _, program := range programs {
			path, err := program.write(each, dir)
			written = append(written, Written{Program: program.name,
				Name: each.Name, Path: path, Err: err})
		}
	}
	return written
}

// header is the words a rendered file opens with. They are the colourway's
// own words, then where it came from, then a note to change the source.
func header(colourway Source) string {
	return colourway.About + "\n\nrendered by colourway from " +
		colourway.Origin + ";\nchange the source, not this file."
}

// writeGhostty writes the terminal theme from the resolved roles. It puts
// claude code's colours in the palette. The selection, surface-1, is the
// one general role the terminal reads that a source may leave to be derived.
func writeGhostty(colourway Source, dir string) (string, error) {
	said, err := claude.Into(Resolve(colourway.Roles).Roles)
	if err != nil {
		return "", err
	}
	theme, err := ghostty.Of(said)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "ghostty", colourway.Name)
	return path, writeFile(path, func(out io.Writer) error {
		return theme.Write(out, header(colourway))
	})
}

// writeNvim writes the editor's colour scheme as lua, from the resolved
// roles. A role the source does not set is derived, so every terminal
// colourway has a scheme.
func writeNvim(colourway Source, dir string) (string, error) {
	scheme, err := nvim.Of(colourway.Name, Resolve(colourway.Roles).Roles)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "nvim", "colors", colourway.Name+".lua")
	return path, writeFile(path, func(out io.Writer) error {
		return scheme.Write(out, header(colourway))
	})
}

// writeVim writes the same scheme for vim. Vim reads vimscript, not lua. It
// gets the same colours on the same groups, those vim can hold.
func writeVim(colourway Source, dir string) (string, error) {
	scheme, err := nvim.Of(colourway.Name, Resolve(colourway.Roles).Roles)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "vim", "colors", colourway.Name+".vim")
	return path, writeFile(path, func(out io.Writer) error {
		return vim.Write(out, scheme, header(colourway))
	})
}

// writeGlamour writes the markdown style from the resolved roles. Json holds
// no comments, so the style has no header. Its source is named in the
// readme.
//
// A colourway has nothing to say but its ink if, once resolved, it has no
// markdown role of its own and no heading other than its ink. That colourway
// is skipped.
func writeGlamour(colourway Source, dir string) (string, error) {
	resolved := Resolve(colourway.Roles)
	if !markdownRoles(resolved) {
		return "", fmt.Errorf(
			"no markdown roles, and no heading but the ink")
	}
	style, err := glamour.Of(resolved.Roles)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "glamour", colourway.Name+".json")
	return path, writeFile(path, func(out io.Writer) error {
		_, err := out.Write(append(style, '\n'))
		return err
	})
}

// markdownRoles reports whether a resolved colourway says anything a style
// needs beyond its ink. That is a markdown role of its own (an md- name), or
// a heading that was set or derived from something other than the ink.
func markdownRoles(resolved Resolved) bool {
	for _, name := range resolved.Roles.Names() {
		if strings.HasPrefix(name, "md-") {
			return true
		}
	}
	_, hasHeading := resolved.Of("heading")
	return hasHeading && resolved.From["heading"] != "ink"
}

// writeFile writes a file whole, making its directory first.
func writeFile(path string, write func(io.Writer) error) error {
	var text bytes.Buffer
	if err := write(&text); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, text.Bytes(), 0o644)
}

// inverted turns every role to its opposite. Each lamp becomes 255 minus
// itself.
func inverted(roles *css.Sheet) *css.Sheet {
	turned := css.New()
	for _, rule := range roles.Rules() {
		lamps, _ := srgb.FromSwatch(rule.Swatch)
		r, g, b := lamps.Bytes()
		turned.Set(rule.Name, srgb.RGB8(255-r, 255-g, 255-b).Swatch())
	}
	return turned
}
