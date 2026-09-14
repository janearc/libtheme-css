// visualdocs is the documentation, shown rather than written: one page
// per idea, a few painted rows and a few lines of text, because this is
// a library about colour and the honest way to explain a colour is to
// put it on the screen. each page can also say the same thing in css,
// and show the go that painted it, which is its own source file, embedded,
// so the three can never drift apart.
//
//	visualdocs                  every page, enter for the next, q to stop
//	visualdocs PAGE             one page, painted
//	visualdocs PAGE --css       the same page as a stylesheet
//	visualdocs PAGE --go        the go that produced it: the page's file
//
// pages: swatch, observer, ok, eye, srgb, ramp, css
package main

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

//go:embed page_*.go
var sources embed.FS

// page is one idea: a name, how to paint it, how to say it in css.
type page struct {
	name string
	show func()
	css  func() string
}

var pages = []page{
	{"swatch", swatchPage, swatchCSS},
	{"observer", observerPage, observerCSS},
	{"ok", okPage, okCSS},
	{"eye", eyePage, eyeCSS},
	{"srgb", srgbPage, srgbCSS},
	{"ramp", rampPage, rampCSS},
	{"css", cssPage, cssCSS},
}

func main() {
	if len(os.Args) > 1 {
		name := os.Args[1]
		mode := ""
		if len(os.Args) > 2 {
			mode = os.Args[2]
		}
		for _, p := range pages {
			if p.name != name {
				continue
			}
			switch mode {
			case "":
				p.show()
			case "--css":
				fmt.Print(p.css())
			case "--go":
				src, err := sources.ReadFile("page_" + name + ".go")
				if err != nil {
					fmt.Fprintln(os.Stderr, "visualdocs:", err)
					os.Exit(1)
				}
				os.Stdout.Write(src)
			default:
				fmt.Fprintf(os.Stderr, "visualdocs: modes are --css and --go, not %q\n", mode)
				os.Exit(2)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "no page called %q; there is %s\n", name, names())
		os.Exit(2)
	}
	in := bufio.NewReader(os.Stdin)
	for i, p := range pages {
		p.show()
		if i == len(pages)-1 {
			return
		}
		fmt.Printf("\n   [%d/%d] enter for %s, q to stop: ", i+1, len(pages), pages[i+1].name)
		line, _ := in.ReadString('\n')
		if strings.TrimSpace(line) == "q" {
			return
		}
		fmt.Println()
	}
}

func names() string {
	n := make([]string, len(pages))
	for i, p := range pages {
		n[i] = p.name
	}
	return strings.Join(n, ", ")
}

// paint is a run of cells in the colour, as the terminal's lamps show
// it: the nearest they can do when the colour is outside their reach.
func paint(s swatch.Swatch, width int) string {
	c, _ := srgb.FromSwatch(s)
	r, g, b := c.Bytes()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, b, strings.Repeat(" ", width))
}

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

func title(t string) { fmt.Printf("== %s\n\n", t) }
func say(lines ...string) {
	for _, l := range lines {
		fmt.Println("   " + l)
	}
}
