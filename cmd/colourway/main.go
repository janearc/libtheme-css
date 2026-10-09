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
//
// render names each file after the source, and heads it with the
// source's opening comment. A general role the source does not set is
// derived, as colourway.Resolve says; a program still missing a role is
// skipped, and said so.
package main

import (
	"fmt"
	"os"

	"github.com/janearc/libtheme-css/colourway"
)

// main is the verb table.
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "render":
		if len(os.Args) != 4 {
			usage()
			os.Exit(2)
		}
		err = render(os.Args[2], os.Args[3])
	case "sheet":
		err = sheet(os.Args[2:])
	case "diff":
		var differed bool
		differed, err = diff(os.Args[2:])
		if err == nil && differed {
			os.Exit(1)
		}
	case "show":
		if len(os.Args) != 3 {
			usage()
			os.Exit(2)
		}
		err = show(os.Args[2])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "colourway:", err)
		os.Exit(1)
	}
}

// usage is the verbs, from the package comment.
func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  colourway render SOURCE DIR
  colourway sheet [--ghostty FILE] [--nvim FILE] [--glamour FILE]`)
}

// render writes every program's file a source can fill, and says what
// it wrote and what it skipped.
func render(path, dir string) error {
	source, err := colourway.Read(path)
	if err != nil {
		return err
	}
	for _, written := range colourway.Render(source, dir) {
		if written.Err != nil {
			fmt.Printf("  %-8s %s: skipped: %v\n", written.Program,
				written.Name, written.Err)
			continue
		}
		fmt.Printf("  %-8s %s\n", written.Program, written.Path)
	}
	return nil
}
