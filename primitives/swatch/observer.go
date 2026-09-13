package swatch

import (
	_ "embed"
	"strconv"
	"strings"
)

// The 1931 observer and the daylight it is most often asked to look at,
// as the CIE published them, so that the numbers below are derived and
// not typed. See data/README.md for where the tables come from.

//go:embed data/ciexyz31_1.csv
var observerCSV string

//go:embed data/d65.csv
var d65CSV string

// table reads "wavelength, v1, v2, ..." lines into a map by wavelength.
func table(csv string) map[int][]float64 {
	out := map[int][]float64{}
	for _, line := range strings.Split(csv, "\n") {
		fields := strings.Split(strings.TrimSpace(line), ",")
		if len(fields) < 2 {
			continue
		}
		nm, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			continue
		}
		row := make([]float64, 0, len(fields)-1)
		for _, f := range fields[1:] {
			v, err := strconv.ParseFloat(strings.TrimSpace(f), 64)
			if err != nil {
				continue
			}
			row = append(row, v)
		}
		out[nm] = row
	}
	return out
}

// observer is the 1931 2-degree colour matching functions: for each
// wavelength, how much of it counts toward X, Y and Z.
var observer = table(observerCSV)

// Illuminant is the swatch of a light given as a spectrum, power by
// wavelength in nanometres. Each wavelength's power is weighted by how
// much the observer counts it toward X, Y and Z, the three sums are taken
// across the visible range, and all three are scaled so that Y is
// exactly 1: the relative form, where the light itself is the white.
func Illuminant(spectrum map[int]float64) Swatch {
	var x, y, z float64
	for nm, power := range spectrum {
		if w, ok := observer[nm]; ok && len(w) == 3 {
			x += power * w[0]
			y += power * w[1]
			z += power * w[2]
		}
	}
	if y == 0 {
		return Black
	}
	return Swatch{x / y, 1, z / y}
}

// d65 is CIE standard illuminant D65: not a real sky but the average of
// noon daylight measured in the 1960s, written down as a spectrum, with
// a nominal colour temperature of 6500 kelvin. It is actually 6504: the
// spectrum was fixed first, then physicists revised a constant in the
// formula that turns temperature into a spectrum, and the number moved
// under it. Nobody redefined the white; the label is slightly wrong
// forever. sRGB and every screen you own assume this white.
func d65() map[int]float64 {
	out := map[int]float64{}
	for nm, row := range table(d65CSV) {
		if len(row) > 0 {
			out[nm] = row[0]
		}
	}
	return out
}

// White is D65 seen by the 1931 observer, scaled so Y is 1. It comes out
// as X 0.95047, Z 1.08883, the numbers every colour library types in by
// hand; here they are computed from the two tables at start-up, so the
// person with fourteen doctorates in choosing a white can check the
// working instead of the typing.
var White = Illuminant(d65())

// Black is no light at all.
var Black = Swatch{}
