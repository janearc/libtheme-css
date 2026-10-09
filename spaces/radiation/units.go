package radiation

import "github.com/janearc/libtheme-css/internal/si"

// in order to accommodate assorted nonvisible wavelengths in our css theme
// library, we need to convert down to our smaller wavelengths, like gamma
// radiation. the constants this takes are the si's own, fixed by definition,
// in internal/si.
const (
	planck = si.Planck
	speed  = si.Speed
	charge = si.Charge
)

// conversion of photon energies. an mev is an energy, not a wavelength, but
// most papers on gamma radiation speak in mev, so these are helpers to move
// back and forth between css and gamma ray burst papers, mostly: FromEnergy
// turns an energy into the wavelength of a photon carrying it.
const (
	Electronvolt = 1.0
	KeV          = 1e3
	MeV          = 1e6
	GeV          = 1e9
	TeV          = 1e12
)

// frequencies (1420 * mhz being the hydrogen line)
const (
	Hz  = 1.0
	KHz = 1e3
	MHz = 1e6
	GHz = 1e9
	THz = 1e12
)

// FromEnergy is the wavelength, in nanometres, of a photon carrying an
// energy in electronvolts.
func FromEnergy(ev float64) float64 { return hc / ev }

// Energy is the energy, in electronvolts, of a photon of a wavelength in
// nanometres: FromEnergy, undone.
func Energy(nm float64) float64 { return hc / nm }

// hc is a photon's energy times its wavelength, in electronvolt nanometres.
const hc = planck * speed / charge * 1e9

// takes Hertz, gives Nanometers
func HertzToNm(hz float64) float64 { return speed * 1e9 / hz }

// takes Nanometers, gives Hertz.
func NmToHertz(nm float64) float64 { return speed * 1e9 / nm }
