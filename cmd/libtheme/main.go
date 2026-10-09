// libtheme is the console end of the library: a way to look at a swatch,
// in every form the library has, painted, so that a change lower down
// shows up as a change on screen. It is a check, not a product.
//
//	libtheme show '#ff6ec7'                one colour, every form, painted
//	libtheme show 'oklch(74% 0.20 345)'    the same, from the polar form
//	libtheme ramp '#160d2b' '#ffa2ff' 24   the line between two colours in
//	                                       oklab and, for contrast, in the
//	                                       lamps, so the mud is visible
//	libtheme known                         everything the library derives
//	                                       unprompted: the visual test
//	libtheme known --css                   the same set as a css sheet,
//	                                       plain, for piping to a file
//	libtheme known --paint                 the sheet with each colour
//	                                       painted beside its rule
//	libtheme roundtrip                     each known colour written as
//	                                       hex and oklch(), read back by
//	                                       the tool's own parser, compared
//	libtheme read FILE                     a colourway as the library sees
//	                                       it: roles painted, ramps drawn
package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/internal/age"
	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// build and built are stamped by game build: the commit, and the
// commit's time. --age prints them.
var build, built = "dev", ""

// main is the verb table.
func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "--age", "version":
		fmt.Println(age.Of("libtheme", build, built, time.Now()))
		return
	case "known":
		mode := ""
		if len(os.Args) > 2 {
			mode = os.Args[2]
		}
		err = known(mode)
	case "roundtrip":
		err = roundtrip()

	case "show":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		err = show(os.Args[2])
	case "read":
		if len(os.Args) < 3 {
			usage()
			os.Exit(2)
		}
		err = read(os.Args[2])
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

// usage is the one line to type when the verb was wrong.
func usage() {
	fmt.Fprintln(os.Stderr, "libtheme known [--css|--paint] | roundtrip |"+
		" show COLOUR | ramp COLOUR COLOUR [steps] | read FILE.css\n"+
		"  COLOUR is #rrggbb or oklch(L% C H)")
}

// parse reads the two spellings a person types: a hex code, or css's
// oklch(). Any other spelling is an error.
func parse(s string) (swatch.Swatch, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "oklch(") && strings.HasSuffix(s, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(s, "oklch("),
			")")
		fields := strings.Fields(inner)
		if len(fields) != 3 {
			return swatch.Black,
				fmt.Errorf("oklch wants three numbers, not %q",
					s)
		}
		var c ok.OKLCH
		var err error
		if c.L, err = percentOrNumber(fields[0]); err != nil {
			return swatch.Black, err
		}
		if c.C, err = strconv.ParseFloat(fields[1], 64); err != nil {
			return swatch.Black, err
		}
		// css's "none" for a hue means there is no hue to speak of.
		// The colour is grey to an eye, so any angle would do. Zero is
		// used.
		if fields[2] == "none" {
			c.H = 0
		} else if c.H, err = strconv.ParseFloat(
			fields[2], 64); err != nil {
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

// The derived white shows a seam here: its chroma in ok is about 9e-5,
// because the fit was normalised to a white rounded to four places and
// the swatch's white is the 1 nm integration. Far below anything an eye
// could see; not zero; and OKLCH.String prints its hue as none.

// percentOrNumber reads "74%" as 0.74 and "0.74" as itself.
func percentOrNumber(s string) (float64, error) {
	if strings.HasSuffix(s, "%") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		return v / 100, err
	}
	return strconv.ParseFloat(s, 64)
}

// painting is whether colour cells are drawn: only into a terminal, and
// not when NO_COLOR asks for none. Piped into a file, a diff or a screen
// reader, the names and numbers say it all, and nothing is drawn.
func painting() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// paint is a cell of the colour, as the terminal's own lamps would show it.
// It uses a 24-bit background. If the colour is out of gamut, it uses the
// nearest the screen can do. When not painting, it is empty.
func paint(s swatch.Swatch, width int) string {
	if !painting() {
		return ""
	}
	c, _ := srgb.FromSwatch(s)
	r, g, b := c.Bytes()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm%s\x1b[0m",
		r, g, b, strings.Repeat(" ", width))
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
		gamut = "out of gamut, clipped"
	}
	fmt.Println(strings.TrimRight(arg+"  "+paint(s, 12), " "))
	fmt.Printf("  xyz     %.5f %.5f %.5f\n", x, y, z)
	fmt.Printf("  oklab   %.4f %.4f %.4f\n", lab.L, lab.A, lab.B)
	fmt.Printf("  oklch   %s\n", lch)
	fmt.Printf("  srgb    %.4f %.4f %.4f  %s\n", c.R, c.G, c.B, gamut)
	fmt.Printf("  hex     %s\n", c.Hex())
	h, v := c.HSL(), c.HSV()
	fmt.Printf("  hsl     %.1f %.3f %.3f\n", h.H, h.S, h.L)
	fmt.Printf("  hsv     %.1f %.3f %.3f\n", v.H, v.S, v.V)
	fmt.Printf("  to white %.3f   to black %.3f   (oklab; eye is %.2f)\n",
		ok.Distance(lab, ok.White), ok.Distance(lab, ok.Black), ok.Eye)
	return nil
}

// ramp draws the line between two colours twice. The first is in oklab,
// which is the line the library uses. The second is a straight line through
// the lamps (srgb), which is what other tools do. This puts the difference
// on screen instead of in an argument.
//
// Both are the same Ramp with a different mixer. That is why the mixer is a
// parameter.
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
	for _, in := range []functions.Mixer{ok.Mix, srgb.Mix} {
		r := functions.Even(in, sa, sb)
		var line strings.Builder
		for _, s := range r.Samples(n) {
			line.WriteString(paint(s, 2))
		}
		fmt.Printf("%-6s %s\n       %s\n",
			in.Name, line.String(), r.String(hex))
	}
	return nil
}

// hex is a swatch written the way a stop in css is.
func hex(s swatch.Swatch) string {
	c, _ := srgb.FromSwatch(s)
	return c.Hex()
}

// The text has no attributes. Dim is unreadable on a dark palette, and
// bold reads differently on every terminal. The output has to read on any
// palette, so the words are plain and only the swatches are painted.
//
// Telling a light terminal from a dark one depends on the terminal, and
// this does not try.
const (
	dim   = ""
	bold  = ""
	plain = ""
)

// heading prints one line saying what the block under it is. It separates
// the output from whatever printed above it.
func heading(text string) { fmt.Printf("-- %s\n", text) }

// paintSheet writes the sheet with each rule's colour painted before it.
func paintSheet(sheet *css.Sheet) {
	rules := sheet.Rules()
	width := 0
	for _, r := range rules {
		if len(r.Name) > width {
			width = len(r.Name)
		}
	}
	fmt.Printf("     :root {\n")
	for _, r := range rules {
		fmt.Printf("  %s   %s--%s:%s%*s %s;  %s/* %s */%s\n",
			paint(r.Swatch,
				2), bold, r.Name, plain, width-len(r.Name), "",
			r.Value, dim, r.Comment, plain)
	}
	fmt.Printf("     }\n")
}

// roundtrip is the translation test. Every known colour is written two
// ways, as hex and as oklch(). Both are read back through parse and the two
// swatches are compared.
//
// Within Exact they are the same colour to arithmetic. Within Eye they are
// the same colour to a person, which is all the oklch spelling, printed to
// three places, can promise. A colourway file is only as good as this trip.
func roundtrip() error {
	type entry struct {
		name string
		s    swatch.Swatch
	}
	list := []entry{
		{"black", swatch.Black}, {"white", swatch.White},
		{"red", srgb.Red.Swatch()},
		{"green", srgb.Green.Swatch()},
		{"blue", srgb.Blue.Swatch()},
	}
	heading("each colour written two ways and read back through the same" +
		" parser.")
	heading("apart: the distance in oklab between the two readings.")
	heading(fmt.Sprintf("under %.0e is arithmetic noise; under %.2g a "+
		"person"+
		" cannot tell", ok.Exact, ok.Eye))
	heading("them apart; over that, they could.")
	fmt.Printf("   %-6s %-8s %-24s %-8s %s\n",
		"", "hex", "oklch", "apart", "verdict")
	failed := false
	for _, e := range list {
		h := hex(e.s)
		l := ok.FromSwatch(e.s).Polar().String()
		fromHex, err := parse(h)
		if err != nil {
			return err
		}
		fromLCH, err := parse(l)
		if err != nil {
			return err
		}
		d := ok.Distance(ok.FromSwatch(fromHex), ok.FromSwatch(fromLCH))
		verdict := "same to a person"
		switch {
		case d <= ok.Exact:
			verdict = "identical"
		case d > ok.Eye:
			verdict = "different to a person"
			failed = true
		}
		fmt.Printf("%s %-6s %-8s %-24s %-8.5f %s\n",
			paint(e.s, 2), e.name, h, l, d, verdict)
	}
	if err := roundtripRadiation(); err != nil {
		return err
	}
	if failed {
		return fmt.Errorf("a round trip lost more than an eye can miss")
	}
	return nil
}

// roundtripRadiation is the same trip for radiation. Each wavelength is
// drawn in false colour, written as hex, read back through parse, and named
// again with Of. It comes back when Of names a wavelength in the same cell
// of the bar.
//
// Profiles drawn onto the same range take the same trip, cell for cell, so
// one bar answers for each range. A wavelength's colour is outside srgb, so
// it clips, and a run of wavelengths can clip to the same hex. Past about
// 690 nm of the render, every wavelength looks the same red.
//
// Nothing here fails. The count says how much a colourway file can carry.
func roundtripRadiation() error {
	fmt.Println()
	heading("radiation: each wavelength drawn, written as hex, read back")
	heading("and named again. x marks one that lands in another cell.")
	for _, group := range byRender(radiation.Profiles) {
		bar, marks, back, err := tripAcross(group[0])
		if err != nil {
			return err
		}
		names := "profiles are"
		if len(group) == 1 {
			names = "profile is"
		}
		heading(fmt.Sprintf("%d %s drawn onto %g to %g nm, so one bar"+
			" answers.", len(group), names, group[0].Render.Lo,
			group[0].Render.Hi))
		fmt.Printf("%s\n%s  %d of %d come back\n", bar, marks, back,
			cells)
	}
	return nil
}

// tripAcross takes one profile's wavelengths through the trip. It returns
// the bar as drawn, a row marking each cell that did not come back, and how
// many did.
func tripAcross(
	profile radiation.Profile,
) (bar, marks string, back int, err error) {
	var drawn, missed strings.Builder
	for i, nm := range across(profile) {
		colour := profile.Swatch(nm)
		drawn.WriteString(paint(colour, 2))
		read, parseErr := parse(hex(colour))
		if parseErr != nil {
			return "", "", 0, parseErr
		}
		named, found := profile.Of(read)
		if found && cellOf(profile, named) == i {
			back++
			missed.WriteString("  ")
		} else {
			missed.WriteString(" x")
		}
	}
	return drawn.String(), missed.String(), back, nil
}

// byRender groups profiles by the range they are drawn onto, in the order
// each range first appears.
func byRender(profiles []radiation.Profile) [][]radiation.Profile {
	var groups [][]radiation.Profile
	for _, profile := range profiles {
		placed := false
		for i, group := range groups {
			if group[0].Render == profile.Render {
				groups[i] = append(group, profile)
				placed = true
				break
			}
		}
		if !placed {
			groups = append(groups, []radiation.Profile{profile})
		}
	}
	return groups
}

// known is everything the library can put on screen without being told a
// colour: the points it defines and the lines between them. It is the
// visual test.
//
// Each entry names where the colour is defined. That is the list this
// library is really keeping.
//
// Black and white are the swatch's, from the observer and the daylight.
// Red, green and blue are srgb's, from the standard's chromaticities. The
// grey line is oklab's, the only line the space defines on its own.
//
// A grey by itself is not on the list. "Grey" is not a colour until you say
// how light, and the line says that better than any one point.
//
// Radiation's profiles are lines too. Each lays one kind of radiation across
// the visible spectrum, so its colours stand for wavelengths no eye sees.
func known(mode string) error {
	type entry struct {
		name, from string
		s          swatch.Swatch
	}
	list := []entry{
		{"black", "swatch: no light", swatch.Black},
		{"white", "swatch: d65 via the observer", swatch.White},
		{"red", "srgb: the red lamp at full", srgb.Red.Swatch()},
		{"green", "srgb: the green lamp at full", srgb.Green.Swatch()},
		{"blue", "srgb: the blue lamp at full", srgb.Blue.Swatch()},
	}
	if mode == "--css" || mode == "--paint" {
		sheet := css.New()
		for _, e := range list {
			sheet.Set(e.name, e.s)
		}
		if mode == "--css" {
			fmt.Print(sheet.String())
			return nil
		}
		heading("the same five as a css sheet. the cell before each " +
			"rule is" +
			" its value.")
		paintSheet(sheet)
		return nil
	}
	heading("everything the library can derive without being told a" +
		" colour,")
	heading("and where each one is defined.")
	for _, e := range list {
		lch := ok.FromSwatch(e.s).Polar()
		c, _ := srgb.FromSwatch(e.s)
		fmt.Printf("%s  %-6s %s   %-24s %s\n",
			paint(e.s, 8), e.name, c.Hex(), lch, e.from)
	}
	// The one line the library defines on its own, drawn with each mixer
	// the spaces supply. The two stops are the same, so the disagreement
	// between the spaces about what a straight line is shows on screen.
	fmt.Println()
	for _, in := range []functions.Mixer{ok.Mix, srgb.Mix} {
		r := functions.Even(in, swatch.Black, swatch.White)
		var line strings.Builder
		for _, s := range r.Samples(24) {
			line.WriteString(paint(s, 2))
		}
		fmt.Printf("%s  black to white\n   %s\n",
			line.String(), r.String(hex))
	}
	knownRadiation()
	return nil
}

// cells is how many cells a radiation bar has, and so how many wavelengths
// a round trip samples.
const cells = 24

// knownRadiation draws each of radiation's profiles as a bar: the
// shortest wavelength at the violet end, the longest at the red.
func knownRadiation() {
	fmt.Println()
	heading("radiation in false colour: each kind laid across the visible")
	heading("spectrum, its shortest wavelength at the violet end.")
	for _, profile := range radiation.Profiles {
		var bar strings.Builder
		for _, nm := range across(profile) {
			bar.WriteString(paint(profile.Swatch(nm), 2))
		}
		fmt.Printf("%s  %s\n   %.3g to %.3g nm\n", bar.String(),
			profile.Name, profile.Source.Lo, profile.Source.Hi)
	}
}

// across is one wavelength for each cell of a profile's bar, from the
// middle of the cell. They are spaced evenly in the logarithm of
// wavelength, as the profile lays them onto the visible.
func across(profile radiation.Profile) []float64 {
	nms := make([]float64, cells)
	span := profile.Source.Hi / profile.Source.Lo
	for i := range nms {
		nms[i] = profile.Source.Lo *
			math.Pow(span, (float64(i)+0.5)/cells)
	}
	return nms
}

// cellOf is the cell of a profile's bar that a wavelength falls in. It is
// -1 for a wavelength outside the profile's source.
func cellOf(profile radiation.Profile, nm float64) int {
	at := math.Log(nm/profile.Source.Lo) /
		math.Log(profile.Source.Hi/profile.Source.Lo)
	if at < 0 || at >= 1 {
		return -1
	}
	return int(at * cells)
}

// read shows a colourway file as the library sees it. It paints the roles.
// Then for each ramp it lists the stops with their positions, and paints
// the ramp sampled across a bar. It shows what the reader found and
// nothing it inferred.
func read(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	cw := css.Read(string(src))
	heading(fmt.Sprintf("%s: %d roles", path, len(cw.Roles.Names())))
	paintSheet(cw.Roles)
	for _, name := range cw.Order {
		r := cw.Ramps[name]
		heading(fmt.Sprintf("ramp %s, %d stops", name, len(r.Stops)))
		for _, st := range r.Stops {
			fmt.Printf("  %5.1f%%  %s  %s\n",
				st.At*100, paint(st.Swatch, 4), hex(st.Swatch))
		}
		var bar strings.Builder
		for _, sw := range r.Samples(48) {
			bar.WriteString(paint(sw, 1))
		}
		fmt.Printf("  %s\n", bar.String())
	}
	return nil
}
