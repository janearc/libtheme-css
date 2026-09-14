// visualdocs is the documentation, shown rather than written: one page
// per idea, a few painted rows and a few lines of text, because this is
// a library about colour and the honest way to explain a colour is to
// put it on the screen. `make visualdocs` pages through all of it with
// enter; a page name shows one.
//
//	visualdocs                  every page, enter for the next, q to stop
//	visualdocs swatch           one page: swatch, observer, ok, eye, srgb,
//	                            ramp, css
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// page is a name and a function that prints it.
type page struct {
	name string
	show func()
}

var pages = []page{
	{"swatch", swatchPage}, {"observer", observerPage}, {"ok", okPage},
	{"eye", eyePage}, {"srgb", srgbPage}, {"ramp", rampPage}, {"css", cssPage},
}

func main() {
	if len(os.Args) > 1 {
		for _, p := range pages {
			if p.name == os.Args[1] {
				p.show()
				return
			}
		}
		fmt.Fprintf(os.Stderr, "no page called %q; there is %s\n", os.Args[1], names())
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

func title(t string) { fmt.Printf("== %s\n\n", t) }
func say(lines ...string) {
	for _, l := range lines {
		fmt.Println("   " + l)
	}
}

func swatchPage() {
	title("swatch: one colour, complete")
	fmt.Printf("   %s black    %s white\n\n", paint(swatch.Black, 6), paint(swatch.White, 6))
	say("a swatch is one colour with nothing in it about the screen showing it",
		"or the eye reading it. it is three numbers, cie xyz from 1931: Y is",
		"how much light, white is 1 and black is 0; X and Z are the rest of",
		"what an eye reports. every other way of writing a colour in this",
		"library is a reading of a swatch, and none of them is a new colour.",
		"the fields are hidden: xyz is where the truth is kept, not where the",
		"arithmetic is done, because a straight line between two xyz points",
		"does not look straight.")
}

func observerPage() {
	title("observer: what 'visible' means here")
	var b strings.Builder
	for nm := 380; nm <= 700; nm += 5 {
		s := swatch.Monochrome(nm)
		// lift each wavelength to full brightness so the hue shows; the
		// clipping to the screen's reach is the honest part
		c, _ := srgb.FromSwatch(s)
		m := c.R
		if c.G > m {
			m = c.G
		}
		if c.B > m {
			m = c.B
		}
		if m > 0 {
			c = srgb.RGB{R: c.R / m, G: c.G / m, B: c.B / m}
		}
		b.WriteString(paint(c.Swatch(), 1))
	}
	fmt.Printf("   %s\n   380 nm%sup to 700 nm\n\n", b.String(), strings.Repeat(" ", 65-12))
	say("the visible spectrum, one cell per five nanometres, as the 1931",
		"observer sees it: seventeen adults in two english labs, matching",
		"lamp light in a dark room, averaged into a table. that table is the",
		"library's whole definition of visible; ultraviolet and infrared",
		"contribute nothing because the table has no row for them. almost",
		"every colour in this bar is outside what a screen can make, so what",
		"you see is the nearest the lamps can do, which is what every screen",
		"has always shown you of a rainbow. the white, d65, is daylight run",
		"through the same table, derived at start-up, not typed.")
}

func okPage() {
	title("ok: the axes an eye agrees with")
	grey := functions.Even(ok.Mix, swatch.Black, swatch.White)
	fmt.Printf("   %s  L: black to white\n", bar(grey, 40))
	var wheel strings.Builder
	for i := 0; i < 40; i++ {
		wheel.WriteString(paint(ok.OKLCH{L: 0.7, C: 0.12, H: float64(i) * 360 / 40}.Rect().Swatch(), 1))
	}
	fmt.Printf("   %s  h: once round the wheel\n", wheel.String())
	var chroma strings.Builder
	for i := 0; i < 40; i++ {
		chroma.WriteString(paint(ok.OKLCH{L: 0.7, C: 0.2 * float64(i) / 39, H: 345}.Rect().Swatch(), 1))
	}
	fmt.Printf("   %s  C: grey out to the pink\n\n", chroma.String())
	say("oklab is the swatch in three numbers an eye agrees with: L, how",
		"light; a and b, which way round the wheel and how far out. its polar",
		"form, oklch, is the one you say aloud: lightness, chroma, hue. the",
		"point of it is distance: equal distances look equally different",
		"wherever they sit, so blending, ramps and 'nearest' all happen here.",
		"it has no primaries; it has a white, a black, the grey line and the",
		"wheel. anything with a hue is defined elsewhere and measured here.")
}

func eyePage() {
	title("eye: precision, set once")
	base := ok.OKLab{L: 0.6, A: 0.1, B: -0.05}
	fmt.Printf("   %s %s  two colours %.3f apart in oklab: the same to a person\n", paint(base.Swatch(), 6), paint(ok.OKLab{L: 0.61, A: 0.1, B: -0.05}.Swatch(), 6), 0.01)
	fmt.Printf("   %s %s  two colours %.3f apart: just about tellable\n", paint(base.Swatch(), 6), paint(ok.OKLab{L: 0.62, A: 0.1, B: -0.05}.Swatch(), 6), ok.Eye)
	fmt.Printf("   %s %s  two colours %.3f apart: different\n\n", paint(base.Swatch(), 6), paint(ok.OKLab{L: 0.7, A: 0.1, B: -0.05}.Swatch(), 6), 0.1)
	say("there are two tolerances in the library and no others. Exact, a",
		"billionth, is arithmetic noise: a round trip, a sum walked in a",
		"different order. Eye, 0.02, is the smallest difference most people",
		"notice side by side, a measurement of people, averaged, and the",
		"reason oklab exists is that 0.02 means the same for a dark blue as",
		"for a pale yellow. every 'are these the same' compares against one",
		"of the two, so precision is a decision made once, not per call.")
}

func srgbPage() {
	title("srgb: the screen's three lamps")
	fmt.Printf("   %s red   %s green   %s blue   %s all three: the white\n\n",
		paint(srgb.Red.Swatch(), 6), paint(srgb.Green.Swatch(), 6), paint(srgb.Blue.Swatch(), 6), paint(srgb.White.Swatch(), 6))
	loud := ok.OKLCH{L: 0.7, C: 0.4, H: 145}.Rect().Swatch()
	_, in := srgb.FromSwatch(loud)
	fmt.Printf("   %s a green greener than the green lamp. in gamut? %v.\n   what you see is the nearest the lamps can do.\n\n", paint(loud, 6), in)
	say("a screen makes colour with three lamps, and srgb is where each lamp",
		"sits on the 1931 diagram: numbers written down for television",
		"phosphors in 1990 and adopted for every computer screen in 1996. the",
		"library types only the standard's four facts, the three lamps and",
		"the curve, and derives the matrix everyone else copies from them",
		"plus the white. hex is this space. so are hsl and hsv, the same",
		"lamps described by angle. a swatch may fall outside the lamps'",
		"reach; every exit here says whether it did.")
}

func rampPage() {
	title("ramp: a function from a number to a colour")
	a, b := srgb.MustHex("#160d2b").Swatch(), srgb.MustHex("#ffa2ff").Swatch()
	fmt.Printf("   %s  in oklab\n", bar(functions.Even(ok.Mix, a, b), 48))
	fmt.Printf("   %s  in srgb, the same two stops\n", bar(functions.Even(srgb.Mix, a, b), 48))
	three := functions.Even(ok.Mix, srgb.MustHex("#2a0f4d").Swatch(), srgb.MustHex("#d94a8c").Swatch(), srgb.MustHex("#fff4c0").Swatch())
	fmt.Printf("   %s  three stops, even\n", bar(three, 48))
	fmt.Printf("   %s  the same three, sampled at eight\n\n", bar(three, 8))
	say("a ramp is stops, each a swatch at a position from 0 to 1, and a",
		"mixer, a named way of blending two swatches that each space supplies.",
		"give it a t, get a swatch. where t comes from is not this library's",
		"business: a canvas, a strip of lamps and a bar of hours each compute",
		"their own and hand it over, which is what lets a sun, a sky and a",
		"stripe share one ramp. containers, not form. sampled at n, a ramp is",
		"what a strip with n lamps or a palette with n entries gets.")
}

func cssPage() {
	title("css: the container everything normalises to")
	sheet := css.New()
	sheet.Set("ground", srgb.MustHex("#160d2b").Swatch())
	sheet.Set("ink", srgb.MustHex("#d8d0e0").Swatch())
	sheet.Set("pink", srgb.MustHex("#ff6ec7").Swatch())
	for _, r := range sheet.Rules() {
		fmt.Printf("   %s  --%s: %s;  /* %s */\n", paint(r.Swatch, 2), r.Name, r.Value, r.Comment)
	}
	fmt.Println()
	say("a sheet is named swatches in order, written as the custom properties",
		"a browser, an editor theme or a lamp's config would read: the hex a",
		"screen can make, and the same colour as oklch beside it, which is the",
		"form that says what it is. a hex code in css is srgb by definition and",
		"a painted cell is srgb by construction, so the bar and the rule",
		"beside it are the same three bytes in two syntaxes; if they ever",
		"disagree, something between here and the screen is lying.")
}
