package colourway

import (
	"bytes"
	"errors"
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
	"github.com/janearc/libtheme-css/internal/plain"
)

// programs are the files Render writes, in order. Each has the program's
// name and a function that writes one colourway under a directory and
// returns the path it wrote.
var programs = []struct {
	name  string
	write func(each rendering, dir string) (string, error)
}{
	{"ghostty", writeGhostty},
	{"nvim", writeNvim},
	{"vim", writeVim},
	{"glamour", writeGlamour},
}

// Written is one file Render made, or the reason it made none. It holds the
// program, the colourway's name, and the path or the error.
//
// An error is one of two kinds. A skip is a program the source cannot fill:
// a role it needs is missing and cannot be derived. Anything else is a file
// that should have been written and was not.
type Written struct {
	Program, Name, Path string
	Err                 error
}

// Skipped is whether the file was left out because the source cannot fill
// the program, as opposed to failing to be written.
func (made Written) Skipped() bool {
	var left skip
	return errors.As(made.Err, &left)
}

// skip is the reason a program was left out of a render.
type skip struct {
	reason error
}

// Error says why the program was left out.
func (left skip) Error() string { return left.reason.Error() }

// Inverted is the source's twin with every colour turned to its opposite. It
// is for a screen whose colours the system inverts. Its name ends in
// -inverted.
//
// The roles are resolved first and the final colours complemented, so a
// role the source leaves to be derived is the opposite of what the source
// derives, byte for byte, and not derived again from the opposite ground.
func (source Source) Inverted() Source {
	return source.twin(Resolve(source.Roles))
}

// twin is the inverted source, made from the source's roles once resolved.
func (source Source) twin(resolved Resolved) Source {
	return Source{Name: source.Name + "-inverted",
		About: source.About + "\n\ninverted, for a screen whose " +
			"colours the system inverts, so it shows as meant.",
		Roles:  inverted(resolved.Roles),
		Origin: source.Origin}
}

// Render writes every file a source can fill under dir. It does so for the
// source and for its inverted twin. It returns one result for each file: it
// was written or it was skipped for a reason or it failed.
//
// Each program renders from the resolved roles, so a general role the source
// does not set is derived. If a role cannot be derived, that file is skipped.
//
// Files that could be filled and were not written are failures, not skips.
// The source's name may not be one a file can have, or the folder may be
// unwritable. Written.Skipped tells the two apart.
func Render(source Source, dir string) []Written {
	written := []Written{}
	nameErr := plain.Name(source.Name)
	normal := Resolve(source.Roles)
	for _, each := range []rendering{newRendering(source, normal),
		newRendering(source.twin(normal), normal.inverted())} {
		for _, program := range programs {
			path, err := "", nameErr
			if err == nil {
				path, err = program.write(each, dir)
			}
			written = append(written, Written{Program: program.name,
				Name: each.source.Name, Path: path, Err: err})
		}
	}
	return written
}

// rendering is a source and the resolved roles it is rendered from. For the
// inverted twin the roles are the normal ones complemented, with the same
// account of where each came from.
//
// It also holds the editor scheme made from those roles, made once, since
// neovim and vim are written from the same one. If it could not be made,
// schemeErr says why.
type rendering struct {
	source    Source
	resolved  Resolved
	scheme    nvim.Scheme
	schemeErr error
}

// newRendering is a source and its resolved roles, with the editor scheme
// they describe.
func newRendering(source Source, resolved Resolved) rendering {
	scheme, err := nvim.Of(source.Name, resolved.Roles)
	return rendering{source, resolved, scheme, err}
}

// editor is the editor scheme, or the reason the source cannot fill the
// editors.
func (each rendering) editor() (scheme nvim.Scheme, err error) {
	if each.schemeErr != nil {
		return scheme, skip{each.schemeErr}
	}
	return each.scheme, nil
}

// inverted is the resolved roles turned to their opposite, each still
// coming from where it came from.
func (resolved Resolved) inverted() Resolved {
	return Resolved{Roles: inverted(resolved.Roles), From: resolved.From}
}

// Scheme is the editor scheme the source describes, from its resolved
// roles. A role the source does not set is derived, so every terminal
// colourway has a scheme. A role that cannot be derived is an error.
func (source Source) Scheme() (nvim.Scheme, error) {
	return nvim.Of(source.Name, Resolve(source.Roles).Roles)
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
func writeGhostty(each rendering, dir string) (string, error) {
	colourway, resolved := each.source, each.resolved
	said, err := claude.Into(resolved.Roles)
	if err != nil {
		return "", skip{err}
	}
	theme, err := ghostty.Of(said)
	if err != nil {
		return "", skip{err}
	}
	path := filepath.Join(dir, "ghostty", colourway.Name)
	return path, writeFile(path, func(out io.Writer) error {
		return theme.Write(out, header(colourway))
	})
}

// writeNvim writes the editor's colour scheme as lua, from the resolved
// roles. A role the source does not set is derived, so every terminal
// colourway has a scheme.
func writeNvim(each rendering, dir string) (string, error) {
	colourway := each.source
	scheme, err := each.editor()
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
func writeVim(each rendering, dir string) (string, error) {
	colourway := each.source
	scheme, err := each.editor()
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
func writeGlamour(each rendering, dir string) (string, error) {
	colourway, resolved := each.source, each.resolved
	if !markdownRoles(resolved) {
		return "", skip{errors.New(
			"no markdown roles, and no heading but the ink")}
	}
	style, err := glamour.Of(resolved.Roles)
	if err != nil {
		return "", skip{err}
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
		turned.Set(rule.Name, ghostty.Complement(rule.Swatch))
	}
	return turned
}
