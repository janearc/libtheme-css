// visualdocs is the documentation, shown rather than written: one page per
// idea, a few painted rows and a few lines of text, because this is a library
// about colour and the honest way to explain a colour is to put it on the
// screen.
//
// each page can also say the same thing in css, and show the go that painted
// it, which is its own source file, embedded, so the three can never drift
// apart.
//
//	visualdocs                  every page, painted, enter for the next
//	visualdocs PAGE             one page, painted
//	visualdocs [PAGE] --css     the same, as a stylesheet
//	visualdocs [PAGE] --go      the go that produced it: the page's own file
//
// pages: swatch, observer, bands, ok, eye, srgb, ramp, css
package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/janearc/libtheme-css/internal/age"
	"github.com/janearc/libtheme-css/internal/host"
	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

//go:embed page_*.go
var sources embed.FS

// page is one idea. It has a name, a function that paints it, and a
// function that says it in css.
type page struct {
	name string
	show func()
	css  func() string
}

var pages = []page{
	{"swatch", swatchPage, swatchCSS},
	{"observer", observerPage, observerCSS},
	{"bands", bandsPage, bandsCSS},
	{"ok", okPage, okCSS},
	{"eye", eyePage, eyeCSS},
	{"srgb", srgbPage, srgbCSS},
	{"ramp", rampPage, rampCSS},
	{"css", cssPage, cssCSS},
}

// build and built are stamped by game build: the commit, and the
// commit's time. --age prints them.
var build, built = "dev", ""

// The exit statuses. They are defined by the shell's convention.
const (
	// succeeded is the status of a call that did what it was asked. It is
	// defined by the shell.
	succeeded = 0
	// failed is the status of a call that could not. It is defined by the
	// shell.
	failed = 1
	// wrongCall is the status of a call the command cannot make sense of.
	// It is defined by the shell.
	wrongCall = 2
)

// afterProgram is where the command's own arguments start in os.Args. It is
// defined by os.Args.
const afterProgram = 1

// program stops the program with a status. It is os.Exit. A test holds a
// mock in its place, so main can run and say what it stopped with.
var program host.Stopper = host.ExitFunc(os.Exit)

// main stops with the status of the call.
func main() {
	program.Exit(run(os.Args[afterProgram:]))
}

// run is the pager: a page name and a mode, or every page in turn. It returns
// the status to stop with.
func run(arguments []string) int {
	if len(arguments) > nothing &&
		(arguments[first] == "--age" || arguments[first] == "version") {
		fmt.Println(age.Of("visualdocs", build, built, time.Now()))
		return succeeded
	}
	// arguments in any order: an optional page name and an optional mode.
	// no page means every page; no mode means paint.
	var name, mode string
	for _, a := range arguments {
		if strings.HasPrefix(a, "--") {
			mode = a
		} else {
			name = a
		}
	}
	if mode != "" && mode != "--css" && mode != "--go" {
		fmt.Fprintf(os.Stderr,
			"visualdocs: modes are --css and --go, not %q\n", mode)
		return wrongCall
	}
	chosen := pages
	if name != "" {
		chosen = nil
		for _, p := range pages {
			if p.name == name {
				chosen = []page{p}
			}
		}
		if chosen == nil {
			fmt.Fprintf(os.Stderr,
				"no page called %q; there is "+
					"%s\n", name, names())
			return wrongCall
		}
	}
	return pageThrough(chosen, mode)
}

// nothing is the count of arguments there are none of. It is defined by
// counting.
const nothing = 0

// first is where the first argument is. It is defined by counting.
const first = 0

// pageThrough shows the pages in the mode, one after the other. With no mode
// it waits between them for enter, and q stops it. It returns the status to
// stop with.
func pageThrough(chosen []page, mode string) int {
	in := bufio.NewReader(os.Stdin)
	for i, p := range chosen {
		if status := showOne(p, mode); status != succeeded {
			return status
		}
		if i == len(chosen)-1 {
			break
		}
		if mode != "" {
			fmt.Println()
			continue
		}
		fmt.Printf("\n   [%d/%d] enter for %s, q to stop: ", i+1,
			len(chosen), chosen[i+1].name)
		line, _ := in.ReadString('\n')
		if strings.TrimSpace(line) == "q" {
			break
		}
		fmt.Println()
	}
	return succeeded
}

// showOne shows one page in the mode: painted, as css, or as its go.
func showOne(shown page, mode string) int {
	switch mode {
	case "--css":
		fmt.Printf("/* %s */\n", shown.name)
		fmt.Print(shown.css())
	case "--go":
		src, err := sources.ReadFile("page_" + shown.name + ".go")
		if err != nil {
			fmt.Fprintln(os.Stderr, "visualdocs:", err)
			return failed
		}
		fmt.Printf("// page_%s.go\n", shown.name)
		os.Stdout.Write(src)
	default:
		shown.show()
	}
	return succeeded
}

// names is every page, for the usage line.
func names() string {
	n := make([]string, len(pages))
	for i, p := range pages {
		n[i] = p.name
	}
	return strings.Join(n, ", ")
}

// paint is a run of cells in the colour, as the terminal's lamps show it.
// If the colour is outside their reach, it is the nearest they can do.
//
// The docs assume two things about the terminal and nothing more. It paints
// a 24-bit background. Its default text colour reads on its default
// background.
//
// Nothing sets a foreground and nothing dims or bolds. The words read on a
// light terminal as well as a dark one, and only the swatches carry colour.
//
// With NO_COLOR set, no escape is written at all. A swatch is a run of
// hashes, its shape without its colour. The hex beside it in the text says
// the colour.
func paint(s swatch.Swatch, width int) string {
	if noColour() {
		return strings.Repeat("#", width)
	}
	c, _ := srgb.FromSwatch(s)
	r, g, b := c.Bytes()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, b,
		strings.Repeat(" ", width))
}

// noColour is the NO_COLOR convention: set to anything, colour is off.
func noColour() bool { return os.Getenv("NO_COLOR") != "" }

// bar is a ramp sampled across a width, painted.
func bar(r functions.Ramp, width int) string {
	var b strings.Builder
	for _, s := range r.Samples(width) {
		b.WriteString(paint(s, 1))
	}
	return b.String()
}

// hexOf is a swatch written the way a stop in css is.
func hexOf(s swatch.Swatch) string {
	c, _ := srgb.FromSwatch(s)
	return c.Hex()
}

// title prints a page's heading.
func title(t string) { fmt.Printf("== %s\n\n", t) }

// say prints lines of text under a heading.
func say(lines ...string) {
	for _, l := range lines {
		fmt.Println("   " + l)
	}
}

// lines splits prose into lines, dropping the first newline.
func lines(prose string) []string {
	return strings.Split(strings.TrimPrefix(prose, "\n"), "\n")
}
