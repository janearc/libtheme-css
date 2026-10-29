// colourway turns a colourway's source, one css sheet, into the files
// each program reads, so nobody writes a theme by hand and nobody who
// only wants a theme has to run anything.
//
//	colourway render SOURCE DIR      every program's file from one source,
//	                                 and the inverted twin of each, under
//	                                 DIR/ghostty, DIR/nvim/colors and
//	                                 DIR/glamour
//	colourway sheet [--ghostty FILE] [--nvim FILE] [--glamour FILE]
//	                                 a source made from programs' files,
//	                                 for a colourway that has none yet
//	colourway diff [--colour] OLD NEW
//	                                 how two colourways differ, role by
//	                                 role, whatever file each is in
//	colourway diff --git PATH OLD OLDHEX OLDMODE NEW NEWHEX NEWMODE
//	                                 the same, as git's diff driver
//	colourway show FILE              each role's hex and oklch, a line
//	                                 each, for git to compare
//	colourway check [--policy FILE] SOURCE...
//	                                 each source's editor scheme held to the
//	                                 reader's measured range, every colour
//	                                 outside it printed; exit 1 if any
//
// render names each file after the source, and heads it with the
// source's opening comment. A general role the source does not set is
// derived, as colourway.Resolve says; a program still missing a role is
// skipped, and said so.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/internal/age"
	"github.com/janearc/libtheme-css/internal/host"
)

// build and built are stamped by game build: the commit, and the
// commit's time. --age prints them.
var build, built = "dev", ""

// streams are where a verb writes: its answer, and its complaints.
type streams struct {
	out, errs io.Writer
}

// The exit statuses. They are defined by the shell's convention.
const (
	// succeeded is the status of a verb that did what it was asked. It is
	// defined by the shell.
	succeeded = 0
	// failed is the status of a verb that could not, or whose answer is
	// no. It is defined by the shell.
	failed = 1
	// wrongCall is the status of a call the command cannot make sense of.
	// It is defined by the shell.
	wrongCall = 2
)

// usageError is a call the command cannot make sense of: no verb, a verb
// nobody has, the wrong number of arguments, an option nobody knows. It is
// the caller's mistake and not the files', so it exits wrongCall and shows
// the usage.
type usageError struct {
	why string
}

// Error says what was wrong with the call.
func (wrong usageError) Error() string { return wrong.why }

// errDiffered is what diff returns when the two colourways differ. Nothing
// is wrong, so nothing is said; the status is failed, as diff's is.
var errDiffered = errors.New("the colourways differ")

// verbs are what the command can be asked to do.
var verbs = map[string]func(arguments []string, to streams) error{
	"--age":   ageOf,
	"version": ageOf,
	"render":  render,
	"sheet":   sheet,
	"diff":    diff,
	"check":   check,
	"show":    show,
}

// afterProgram is where the command's own arguments start in os.Args. It is
// defined by os.Args.
const afterProgram = 1

// program stops the program with a status. It is os.Exit. A test holds a
// mock in its place, so main can run and say what it stopped with.
var program host.Stopper = host.ExitFunc(os.Exit)

// main runs the verb and exits with its status.
func main() {
	program.Exit(run(os.Args[afterProgram:], streams{os.Stdout, os.Stderr}))
}

// run is the command: the verb's name, then its arguments. It returns the
// exit status.
func run(arguments []string, to streams) int {
	if len(arguments) == nothing {
		return wrongCallStatus(to, usageError{"no verb"})
	}
	name, rest := arguments[verbAt], arguments[argumentsFrom:]
	verb, found := verbs[name]
	if !found {
		return wrongCallStatus(to, usageError{"no verb " + name})
	}
	err := verb(rest, to)
	var wrong usageError
	switch {
	case err == nil:
		return succeeded
	case errors.As(err, &wrong):
		return wrongCallStatus(to, wrong)
	case errors.Is(err, errDiffered):
		return failed
	}
	fmt.Fprintln(to.errs, "colourway:", err)
	return failed
}

const (
	// verbAt is where the verb is in a call. It is defined by the usage.
	verbAt = 0
	// argumentsFrom is where its arguments start. It is defined by the
	// usage.
	argumentsFrom = 1
)

// wrongCallStatus says what was wrong with a call, shows the usage, and
// returns the status for it.
func wrongCallStatus(to streams, wrong usageError) int {
	fmt.Fprintln(to.errs, "colourway:", wrong)
	usage(to.errs)
	return wrongCall
}

// ageOf prints which commit this command was built from.
func ageOf(_ []string, to streams) error {
	said := age.Of("colourway", build, built, time.Now())
	fmt.Fprintln(to.out, said)
	return nil
}

// usage is the verbs, from the package comment.
func usage(out io.Writer) {
	fmt.Fprintln(out, `usage:
  colourway render SOURCE DIR
  colourway sheet [--ghostty FILE] [--nvim FILE] [--glamour FILE]
  colourway diff [--colour] OLD NEW
  colourway show FILE
  colourway check [--policy FILE] SOURCE...`)
}

// errNothingRendered is what render returns when every file was left out,
// or none could be written: there is nothing to show for the call.
var errNothingRendered = errors.New("nothing could be rendered")

// render writes every program's file that a source can fill. It prints what
// it wrote, what it skipped because the source cannot fill the program, and
// what it failed to write.
//
// It fails if any file failed to write, and if there is no file at all to
// show for it. A skip alone is not a failure: a source made for a terminal
// has no business filling a markdown style.
func render(arguments []string, to streams) error {
	if len(arguments) != renderArguments {
		return usageError{"render wants a source and a folder"}
	}
	source, err := colourway.Read(arguments[sourceAt])
	if err != nil {
		return err
	}
	var wrote, lost int
	for _, made := range colourway.Render(source, arguments[folderAt]) {
		switch {
		case made.Err == nil:
			wrote++
			fmt.Fprintf(to.out, "  %-8s %s\n",
				made.Program, made.Path)
		case made.Skipped():
			fmt.Fprintf(to.out, "  %-8s %s: skipped: %v\n",
				made.Program, made.Name, made.Err)
		default:
			lost++
			fmt.Fprintf(to.out, "  %-8s %s: failed: %v\n",
				made.Program, made.Name, made.Err)
		}
	}
	switch {
	case lost > nothing:
		return fmt.Errorf("%d files could not be written", lost)
	case wrote == nothing:
		return errNothingRendered
	}
	return nil
}

const (
	// renderArguments is how many arguments render takes. It is defined
	// by the usage.
	renderArguments = 2
	// sourceAt is where the source is. It is defined by the usage.
	sourceAt = 0
	// folderAt is where the folder to write under is. It is defined by
	// the usage.
	folderAt = 1
)
