package bands

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// Share is the fraction of light passed in each of a run of bands: what a
// filter lets through, or what a surface sends back. It is not light. Light
// arriving and a share of it are different things, so they are different
// types, and Lit is the one way to put them together.
//
// Outside its run a share passes nothing, unless it was made Even, which
// passes the same part in every band. The zero value passes nothing.
//
// Deprecated: use celestial's bands.Share.
type Share struct {
	first  Band
	part   []float64
	beyond float64 // the part passed outside the run
}

// whole is whether a part is a fraction a share can be: 0 to 1, and a
// number. A surface that sends back more than arrives is a lamp.
func whole(part float64) error {
	if part >= 0 && part <= 1 {
		return nil
	}
	return fmt.Errorf("a share is a fraction between 0 and 1, and %v "+
		"is not one", part)
}

// Even is a share that passes the same part of the light in every band:
// a grey filter, or a perfectly grey surface.
//
// Deprecated: use celestial's bands.Even, which returns a *Share.
func Even(part float64) (Share, error) {
	if err := whole(part); err != nil {
		return Share{}, err
	}
	return Share{beyond: part}, nil
}

// Shares is a share given as the part passed at wavelengths in
// nanometres, as measured: a reflectance or transmittance curve.
//
// A band's part is the mean of the wavelengths given in it, since a share
// is a fraction and fractions do not add. A band inside the run with none
// given takes the straight line between its neighbours, read at its
// centre. Outside the run nothing passes.
//
// Deprecated: use celestial's bands.Shares, which returns a *Share.
func Shares(spectrum map[int]float64) (Share, error) {
	if len(spectrum) == 0 {
		return Share{}, errors.New("a share needs at least one " +
			"wavelength and the part passed there")
	}
	order := make([]int, 0, len(spectrum))
	for nm, part := range spectrum {
		if nm <= 0 {
			return Share{}, fmt.Errorf("a share at %d nm: a "+
				"wavelength is longer than nothing", nm)
		}
		if err := whole(part); err != nil {
			return Share{}, fmt.Errorf("at %d nm: %w", nm, err)
		}
		order = append(order, nm)
	}
	sort.Ints(order)
	step := grid(PerOctave)
	first := step.of(float64(order[0]))
	last := step.of(float64(order[len(order)-1]))
	share := Share{first: first,
		part: make([]float64, last-first+1)}
	for band := first; band <= last; band++ {
		share.part[band-first] = partIn(step, band, order, spectrum)
	}
	return share, nil
}

// partIn is one band's part: the mean of the wavelengths given in it, or
// the line between its neighbours at its centre when none are.
func partIn(
	step grid, band Band, order []int, spectrum map[int]float64,
) float64 {
	var sum float64
	count := 0
	for _, nm := range order {
		if step.of(float64(nm)) == band {
			sum += spectrum[nm]
			count++
		}
	}
	if count > 0 {
		return sum / float64(count)
	}
	lo, hi := step.edges(band)
	centre := math.Sqrt(lo * hi)
	above := sort.Search(len(order), func(index int) bool {
		return float64(order[index]) > centre
	})
	below, after := order[above-1], order[above]
	along := (centre - float64(below)) / float64(after-below)
	return spectrum[below] + (spectrum[after]-spectrum[below])*along
}

// Part is the fraction the share passes in one band.
func (share Share) Part(band Band) float64 {
	index := int(band - share.first)
	if index < 0 || index >= len(share.part) {
		return share.beyond
	}
	return share.part[index]
}

// Lit is light after a share of it: in every band, the light's power
// times the part passed there.
//
// Deprecated: use celestial's bands.Lit, which takes a *Share.
func Lit(light Light, share Share) Light {
	out := Light{g: light.g, first: light.first,
		power: make([]float64, len(light.power))}
	for index, power := range light.power {
		out.power[index] = power * share.Part(light.first+Band(index))
	}
	return out
}
