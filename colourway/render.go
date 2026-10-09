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

// programs are the files Render writes, in order: each program's name
// and how it writes one colourway under a directory, saying where.
var programs = []struct {
	name  string
	write func(colourway Source, dir string) (string, error)
}{
	{"ghostty", writeGhostty},
	{"nvim", writeNvim},
	{"vim", writeVim},
	{"glamour", writeGlamour},
}

// Written is one file Render made, or the reason it made none: the
// program, the colourway's name, and the path or the error.
type Written struct {
	Program, Name, Path string
	Err                 error
}

// Inverted is the source's twin, every role turned to its opposite, for
// a screen whose colours the system inverts.
func (source Source) Inverted() Source {
	return Source{Name: source.Name + "-inverted",
		About: source.About + "\n\ninverted, for a screen whose " +
			"colours the system inverts, so it shows as meant.",
		Roles: inverted(source.Roles), Origin: source.Origin}
}

// Render writes every program's file a source can fill, under dir, and
// the inverted twin of each: what it wrote, and what it skipped and why.
// Each program renders from the roles resolved, so a general role the
// source does not set is derived; one that nothing derives is skipped.
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

// header is the words a rendered file opens with: the colourway's own,
// then where it came from.
func header(colourway Source) string {
	return colourway.About + "\n\nrendered by colourway from " +
		colourway.Origin + ";\nchange the source, not this file."
}

// writeGhostty writes the terminal theme, claude code's colours in its
// palette, from the roles resolved: the selection, surface-1, is the one
// general role the terminal reads that a source may leave to be derived.
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

// writeNvim writes the editor's colour scheme, from the colourway's roles
// resolved: a role it does not set is derived, so every terminal
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

// writeVim writes the same scheme for vim, which reads vimscript and not
// lua: the same colours on the same groups, those vim can hold.
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

// writeGlamour writes the markdown style, from the colourway's roles
// resolved. json holds no comments, so the style carries no header; its
// source is named in the readme.
//
// A colourway that, resolved, has no markdown role of its own and no
// heading other than its ink has nothing to say but the ink, and is
// skipped.
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

// markdownRoles is whether a colourway, resolved, says anything a style
// needs beyond its ink: a markdown role of its own, or a heading set or
// derived from something other than the ink.
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

// inverted is every role turned to its opposite, each lamp 255 less
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
