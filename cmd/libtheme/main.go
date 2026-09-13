// libtheme is the console end of the library: a way to look at a swatch,
// in every form the library has, painted, so that a change lower down
// shows up as a change on screen. It is a check, not a product.
//
//	libtheme show '#ff6ec7'                 one colour, every form, painted
//	libtheme show 'oklch(74% 0.20 345)'     the same, from the polar form
//	libtheme ramp '#160d2b' '#ffa2ff' 24    the line between two colours,
//	                                        drawn in oklab and, for contrast,
//	                                        in the lamps, so the mud is visible
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "show":
		err = show(os.Args[2])
	case "ramp":
		n := 24
		if len(os.Args) > 4 {
			n, err = strconv.Atoi(os.Args[4])
			if err != nil {
				break
			}
		}
		if len(os.Args) < 4 {
			usage()
			os.Exit(2)
		}
		err = ramp(os.Args[2], os.Args[3], n)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "libtheme:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "libtheme show COLOUR | ramp COLOUR COLOUR [steps]\n  COLOUR is #rrggbb or oklch(L% C H)")
}

// parse reads the two spellings a person types: a hex code, or css's
// oklch(). Everything else is for later.
func parse(s string) (swatch.Swatch, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "oklch(") && strings.HasSuffix(s, ")") {
		fields := strings.Fields(strings.TrimSuffix(strings.TrimPrefix(s, "oklch("), ")"))
		if len(fields) != 3 {
			return swatch.Black, fmt.Errorf("oklch wants three numbers, not %q", s)
		}
		var c ok.OKLCH
		var err error
		if c.L, err = percentOrNumber(fields[0]); err != nil {
			return swatch.Black, err
		}
		if c.C, err = strconv.ParseFloat(fields[1], 64); err != nil {
			return swatch.Black, err
		}
		if c.H, err = strconv.ParseFloat(fields[2], 64); err != nil {
			return swatch.Black, err
		}
		return c.Rect().Swatch(), nil
	}
	c, err := srgb.FromHex(s)
	if err != nil {
		return swatch.Black, err
	}
	return c.Swatch(), nil
}

// percentOrNumber reads "74%" as 0.74 and "0.74" as itself.
func percentOrNumber(s string) (float64, error) {
	if strings.HasSuffix(s, "%") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		return v / 100, err
	}
	return strconv.ParseFloat(s, 64)
}

// paint is a cell of the colour, as the terminal's own lamps would show
// it: 24-bit background, the nearest the screen can do if it is out of
// gamut.
func paint(s swatch.Swatch, width int) string {
	c, _ := srgb.FromSwatch(s)
	r, g, b := c.Bytes()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, b, strings.Repeat(" ", width))
}

// show prints one swatch in every form the library has, and paints it.
func show(arg string) error {
	s, err := parse(arg)
	if err != nil {
		return err
	}
	x, y, z := s.XYZ()
	lab := ok.FromSwatch(s)
	lch := lab.Polar()
	c, in := srgb.FromSwatch(s)
	gamut := "in gamut"
	if !in {
		gamut = "OUT OF GAMUT, clipped"
	}
	fmt.Printf("%s  %s\n", paint(s, 12), arg)
	fmt.Printf("  xyz     %.5f %.5f %.5f\n", x, y, z)
	fmt.Printf("  oklab   %.4f %.4f %.4f\n", lab.L, lab.A, lab.B)
	fmt.Printf("  oklch   %.0f%% %.3f %.1f\n", lch.L*100, lch.C, lch.H)
	fmt.Printf("  srgb    %.4f %.4f %.4f  %s\n", c.R, c.G, c.B, gamut)
	fmt.Printf("  hex     %s\n", c.Hex())
	h, v := c.HSL(), c.HSV()
	fmt.Printf("  hsl     %.1f %.3f %.3f\n", h.H, h.S, h.L)
	fmt.Printf("  hsv     %.1f %.3f %.3f\n", v.H, v.S, v.V)
	fmt.Printf("  to white %.3f   to black %.3f   (oklab; eye is %.2f)\n",
		ok.Distance(lab, ok.White), ok.Distance(lab, ok.Black), ok.Eye)
	return nil
}

// ramp draws the line between two colours twice: once in oklab, which
// is the line the library uses, and once as a straight line in the
// lamps, which is what every other tool does, so the difference is on
// screen and not in an argument.
func ramp(a, b string, n int) error {
	sa, err := parse(a)
	if err != nil {
		return err
	}
	sb, err := parse(b)
	if err != nil {
		return err
	}
	if n < 2 {
		n = 2
	}
	la, lb := ok.FromSwatch(sa), ok.FromSwatch(sb)
	ca, _ := srgb.FromSwatch(sa)
	cb, _ := srgb.FromSwatch(sb)
	var inOK, inLamps strings.Builder
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		mid := ok.OKLab{L: la.L + (lb.L-la.L)*t, A: la.A + (lb.A-la.A)*t, B: la.B + (lb.B-la.B)*t}
		inOK.WriteString(paint(mid.Swatch(), 2))
		lamps := srgb.RGB{R: ca.R + (cb.R-ca.R)*t, G: ca.G + (cb.G-ca.G)*t, B: ca.B + (cb.B-ca.B)*t}
		inLamps.WriteString(paint(lamps.Swatch(), 2))
	}
	fmt.Printf("oklab  %s\n", inOK.String())
	fmt.Printf("lamps  %s\n", inLamps.String())
	return nil
}
