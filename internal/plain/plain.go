// Package plain makes text safe to put in a generated file: a name that
// sits inside quotes and a comment, and words that sit in a comment.
//
// Editors and terminals load generated files, so text from a file name or a
// source file's opening comment must not end a comment or a string and
// begin a command. A line break does that in every format libtheme writes.
//
// Readers differ on what breaks a line: lua ends one at a carriage return,
// vim at a newline. This package treats them alike.
package plain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// newline is the one line end the lines of a comment are made with.
const newline = '\n'

// Name checks a colourway's name. The name is written into quoted strings,
// comments and file names, so it must be valid text.
//
// Every character is graphic, as unicode.IsGraphic has it: a letter, mark,
// number, punctuation, symbol or space. Control characters and line breaks
// are refused. So are invisible characters, such as a zero width space.
//
// It cannot be a path. It is not empty, and not a dot or dot-dot. It has no
// slash or backslash. Quotes are fine; each writer quotes for its own
// language.
func Name(name string) error {
	switch {
	case name == "":
		return errors.New("a name cannot be empty")
	case name == "." || name == "..":
		return fmt.Errorf("%q is a place, not a name", name)
	case !utf8.ValidString(name):
		return errors.New("a name must be valid text")
	case strings.ContainsAny(name, `/\`):
		return fmt.Errorf("%q holds a slash or a backslash", name)
	}
	for _, each := range name {
		if !unicode.IsGraphic(each) {
			return fmt.Errorf("%q holds %U, which is a control or "+
				"invisible character, not text", name, each)
		}
	}
	return nil
}

// Comment is words as a comment: each line after the mark and a space,
// trailing spaces taken off, and a newline ending every line. The mark is
// the language's, such as "--" or "#".
//
// Every kind of line end (newline, carriage return, the pair, vertical
// tab, form feed, next line, the unicode line and paragraph separators)
// ends a comment line, so the words cannot leave the comment.
//
// Any other control character but a tab is dropped. The words are trimmed
// first.
func Comment(mark, words string) string {
	var text strings.Builder
	for _, line := range lines(words) {
		text.WriteString(strings.TrimRight(mark+" "+line, " "))
		text.WriteByte(newline)
	}
	return text.String()
}

// lines is words as lines with no line end and no control character but
// the tab in them.
func lines(words string) []string {
	pair := strings.NewReplacer("\r\n", string(newline))
	single := pair.Replace(strings.TrimSpace(words))
	return strings.Split(strings.Map(oneKind, single), string(newline))
}

// drop is what strings.Map is given to drop a character. It is defined by
// strings.Map.
const drop = -1

// oneKind turns every line end into the newline, drops the other control
// characters but the tab, and leaves the rest.
func oneKind(each rune) rune {
	switch {
	case breaks(each):
		return newline
	case unicode.IsControl(each) && each != '\t':
		return drop
	}
	return each
}

// breaks is whether a character ends a line for some reader.
func breaks(each rune) bool {
	switch each {
	case '\n', '\r', '\v', '\f', '\u0085', '\u2028', '\u2029':
		return true
	}
	return false
}
