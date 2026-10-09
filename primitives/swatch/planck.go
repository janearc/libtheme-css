package swatch

import (
	"math"

	"github.com/janearc/libtheme-css/internal/si"
)

// Planckian is the swatch of a black body at a temperature in kelvin: the
// colour a lamp means when it gives a colour temperature.
//
// The spectrum comes from Blackbody, and Illuminant turns it into a swatch.
// So the locus comes from the same tables as White and is never typed in.
// Below a few hundred kelvin there is almost no visible light, and the
// result is black.
func Planckian(kelvin float64) Swatch {
	if kelvin <= 0 {
		return Black
	}
	return Illuminant(Blackbody(kelvin))
}

// Blackbody is the spectrum of a black body at a temperature in kelvin, by
// Planck's law, at every wavelength the observer has a row for: power by
// wavelength in nanometres. At zero kelvin or below it is empty.
//
// The power is relative. Planck's leading 2hc squared is left out, since it
// is the same at every wavelength and only proportions are asked for here.
// Wavelengths the observer does not list are not in the map, so this is a
// lamp as an eye meets it, not as a thermometer would.
func Blackbody(kelvin float64) map[int]float64 {
	if kelvin <= 0 {
		return map[int]float64{}
	}
	const (
		h = si.Planck
		c = si.Speed
		k = si.Boltzmann
	)
	spectrum := make(map[int]float64, len(observer))
	for nm := range observer {
		wl := float64(nm) * 1e-9
		spectrum[nm] = 1 / (math.Pow(wl,
			5) * (math.Exp(h*c/(wl*k*kelvin)) - 1))
	}
	return spectrum
}
