package swatch

import (
	_ "embed"
	"sort"
	"strconv"
	"strings"
)

// The CIE 1931 observer and D65 daylight, as the CIE published them. The
// numbers below are derived from these tables, not typed. data/README.md
// says where the tables come from.

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

// observer is the CIE 1931 2-degree standard observer: for each
// wavelength, how much of it counts toward X, Y and Z. It is the
// library's definition of "visible", and it is worth being exact about
// what kind of definition that is.
//
// Visibility is not a property of light. Light has a wavelength; whether anyone
// sees it is a property of the eye looking.
//
// This table is a model of one eye: an average of the matches made by seventeen
// adults with normal colour vision, in 1928-1931, looking at a small patch two
// degrees wide (the width of a thumbnail at arm's length, which lands on the
// fovea), at daylight brightness.
//
// For that average person, in those conditions, wavelengths outside roughly
// 380 to 780 nanometres produce no response. The table shows this by tending
// to zero at its ends, and it stops at 360 and 830.
//
// A different observer, a wider field, a dim room, or an eye that is not
// average would give a different table. This library would accept it in the
// same shape. Every range the library uses downstream comes from this one,
// and it is stated once, here.
var observer = table(observerCSV)

// Observed is the shortest and the longest wavelength, in nanometres, the
// observer has a row for: the library's range of visible, stated once, in
// the table, and inherited by every range downstream.
func Observed() (lo, hi int) {
	first := true
	for nm := range observer {
		if first || nm < lo {
			lo = nm
		}
		if first || nm > hi {
			hi = nm
		}
		first = false
	}
	return lo, hi
}

// Illuminant is the swatch of a light given as a spectrum: power by wavelength
// in nanometres. Each wavelength's power is weighted by how much the observer
// counts it toward X, Y and Z, and the three weighted sums are taken over every
// wavelength the observer has a row for.
//
// Power at wavelengths the observer does not list contributes nothing, which is
// the visible range being applied, not a limit of the spectrum. The three sums
// are then scaled so that Y is exactly 1: the relative form, in which this
// light is, by definition, the white.
//
// The sums are taken from the shortest wavelength to the longest, always.
// Adding floating-point numbers in a different order gives a different last
// bit, and a map is walked in a different order every time.
//
// A fixed order means the same light gives the same swatch on every call.
// Nothing at eight bits a channel could see the difference, but the same
// question should have the same answer.
func Illuminant(spectrum map[int]float64) Swatch {
	order := make([]int, 0, len(spectrum))
	for nm := range spectrum {
		order = append(order, nm)
	}
	sort.Ints(order)
	var x, y, z float64
	for _, nm := range order {
		if w, ok := observer[nm]; ok && len(w) == 3 {
			x += spectrum[nm] * w[0]
			y += spectrum[nm] * w[1]
			z += spectrum[nm] * w[2]
		}
	}
	if y == 0 {
		return Black
	}
	return Swatch{x / y, 1, z / y}
}

// Monochrome is light of one wavelength at unit power, as the observer sees it:
// the observer's own row, as a swatch. Outside the table it is no light.
//
// Drawn through this from 380 to 780, the spectrum is the visible range as
// the seventeen observers saw it. Most of it is outside what any screen can
// make, so every screen clips it.
func Monochrome(nm int) Swatch {
	w, ok := observer[nm]
	if !ok || len(w) != 3 {
		return Black
	}
	return Swatch{w[0], w[1], w[2]}
}

// D65 is CIE standard illuminant D65: not a real sky but the average of noon
// daylight measured in the 1960s, written down as a spectrum, with a nominal
// colour temperature of 6500 kelvin. It returns relative power by wavelength
// in nanometres, 300 to 830, as published. Each call returns a new map.
//
// It is actually 6504. The spectrum was fixed first. Then physicists revised a
// constant in the formula that turns temperature into a spectrum, and the
// number moved under it. Nobody redefined the white; the label is slightly
// wrong forever. sRGB and every screen you own assume this white.
func D65() map[int]float64 {
	out := map[int]float64{}
	for nm, row := range table(d65CSV) {
		if len(row) > 0 {
			out[nm] = row[0]
		}
	}
	return out
}

// White is D65 as seen by the observer above, scaled so Y is 1. It comes
// out as X 0.95047, Z 1.08883, the two numbers most colour libraries
// type in; here they are computed from the two published tables at
// start-up, so the derivation can be checked rather than the typing.
var White = Illuminant(D65())

// Black is no light at all.
var Black = Swatch{}
