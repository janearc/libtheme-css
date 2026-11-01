// Package si holds the physical constants libtheme uses. Each is one of the
// si's defining constants, fixed by definition since 2019. They are exact,
// not measured, and each is typed here once.
//
// Deprecated: libtheme's physics lives in celestial now. Use
// github.com/janearc/celestial/si, where Gravitation is a function read
// from its research data. This package goes in libtheme-css v0.5.0.
package si

// The constants.
const (
	Planck    = 6.62607015e-34  // joule seconds
	Speed     = 299792458.0     // of light, metres a second
	Charge    = 1.602176634e-19 // coulombs, the charge of one electron
	Boltzmann = 1.380649e-23    // joules a kelvin
)
