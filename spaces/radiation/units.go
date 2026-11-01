package radiation

import (
	"github.com/janearc/libtheme-css/primitives/length"
	"github.com/janearc/libtheme-css/primitives/si"
)

// in order to accommodate assorted nonvisible wavelengths in our css theme
// library, we need to convert down to our smaller wavelengths, like gamma
// radiation. the constants this takes are the si's own, fixed by definition,
// in primitives/si.
const (
	planck = si.Planck
	speed  = si.Speed
	charge = si.Charge
)

// the units radiation is spoken in. each has one name, and its multiples
// come from the si's prefixes: a keV is si.Kilo(Electronvolt), the
// hydrogen line is 1420.405751 * si.Mega(Hertz).
//
// an eV is an energy, not a wavelength, but most papers on gamma radiation
// speak in eV, so FromEnergy and Energy move between the two. wavelengths
// are in nanometres, and lengths in metres go through MetresToNm.
const (
	// Deprecated: use celestial's radiation.Electronvolt.
	Electronvolt = 1.0
	// Deprecated: use celestial's radiation.Hertz.
	Hertz = 1.0
)

// FromEnergy is the wavelength, in nanometres, of a photon carrying an
// energy in electronvolts.
//
// Deprecated: use celestial's radiation.FromEnergy.
func FromEnergy(ev float64) float64 { return hc / ev }

// Energy is the energy, in electronvolts, of a photon of a wavelength in
// nanometres: FromEnergy, undone.
//
// Deprecated: use celestial's radiation.Energy.
func Energy(nm float64) float64 { return hc / nm }

// nanometre is one nanometre, in metres.
var nanometre = si.Nano(length.Metre)

// hc is a photon's energy times its wavelength, in electronvolt nanometres.
var hc = planck * speed / charge / nanometre

// MetresToNm is a length in metres as nanometres, the unit every range
// here is in: MetresToNm(3.4 * si.Micro(length.Metre)) is WISE's first
// band, 3400 nm.
//
// Deprecated: use celestial's radiation.MetresToNm.
func MetresToNm(m float64) float64 { return m / nanometre }

// NmToMetres is nanometres as metres: MetresToNm, undone.
//
// Deprecated: use celestial's radiation.NmToMetres.
func NmToMetres(nm float64) float64 { return nm * nanometre }

// takes Hertz, gives Nanometers
//
// Deprecated: use celestial's radiation.HertzToNm.
func HertzToNm(hz float64) float64 { return MetresToNm(speed / hz) }

// takes Nanometers, gives Hertz.
//
// Deprecated: use celestial's radiation.NmToHertz.
func NmToHertz(nm float64) float64 { return speed / NmToMetres(nm) }
