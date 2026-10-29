package si

// The si's decimal prefixes, from the SI Brochure, 9th edition (BIPM,
// 2019), table 7. A unit has one name; its multiples are these functions
// of it, so a kiloparsec is Kilo(length.Parsec) and a keV is
// Kilo(radiation.Electronvolt). Nothing anywhere names a kilo-anything.
const (
	femto = 1e-15
	pico  = 1e-12
	nano  = 1e-9
	micro = 1e-6
	milli = 1e-3
	centi = 1e-2
	kilo  = 1e3
	mega  = 1e6
	giga  = 1e9
	tera  = 1e12
	peta  = 1e15
	exa   = 1e18
)

// Femto is a thousand-million-millionth of the unit.
func Femto(unit float64) float64 { return unit * femto }

// Pico is a million-millionth of the unit.
func Pico(unit float64) float64 { return unit * pico }

// Nano is a thousand-millionth of the unit.
func Nano(unit float64) float64 { return unit * nano }

// Micro is a millionth of the unit.
func Micro(unit float64) float64 { return unit * micro }

// Milli is a thousandth of the unit.
func Milli(unit float64) float64 { return unit * milli }

// Centi is a hundredth of the unit.
func Centi(unit float64) float64 { return unit * centi }

// Kilo is a thousand of the unit.
func Kilo(unit float64) float64 { return unit * kilo }

// Mega is a million of the unit.
func Mega(unit float64) float64 { return unit * mega }

// Giga is a thousand million of the unit.
func Giga(unit float64) float64 { return unit * giga }

// Tera is a million million of the unit.
func Tera(unit float64) float64 { return unit * tera }

// Peta is a thousand million million of the unit. IceCube's neutrinos
// are measured in PeV.
func Peta(unit float64) float64 { return unit * peta }

// Exa is a million million million of the unit.
func Exa(unit float64) float64 { return unit * exa }
