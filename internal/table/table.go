// Package table reads the library's embedded data tables: lines of a
// whole-nanometre wavelength and then some numbers, one line to a
// nanometre.
//
// The tables are the CIE's published data, and every colour the library
// converts is worked out from them. So a row that goes missing is not a
// small fault: it changes every colour. The reader refuses a table it
// cannot read whole, and says which line.
package table

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// parse reads lines of a wavelength and its values into a map by wavelength.
//
// A line that is not blank is a whole-number wavelength and exactly values
// numbers after it. Each number is finite and not negative.
// Each wavelength is one nanometre above the one before.
// A line that breaks any of this is an error that names the line.
func parse(csv string, values int) (map[int][]float64, error) {
	out := map[int][]float64{}
	var previous int
	for place, line := range strings.Split(csv, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		at, row, err := parseRow(line, values)
		if err == nil && len(out) != nothing {
			err = follows(at, previous)
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", place+lineOne,
				err)
		}
		out[at], previous = row, at
	}
	if len(out) == nothing {
		return nil, errors.New("the table has no rows")
	}
	return out, nil
}

// follows is whether a row at one wavelength may come after the row at
// another: it must be the next nanometre.
func follows(at, previous int) error {
	if at == previous+nmStep {
		return nil
	}
	return fmt.Errorf("%d nm after %d nm; rows must rise by %d nm",
		at, previous, nmStep)
}

// parseRow reads one line: a wavelength and values numbers after it.
func parseRow(line string, values int) (at int, row []float64,
	err error) {
	fields := strings.Split(line, ",")
	if len(fields) != values+wavelengthField {
		return at, nil, fmt.Errorf("%d fields, want a wavelength "+
			"and %d values", len(fields), values)
	}
	first, rest := fields[wavelengthAt], fields[wavelengthField:]
	if at, err = strconv.Atoi(strings.TrimSpace(first)); err != nil {
		return at, nil, fmt.Errorf("the wavelength: %w", err)
	}
	for _, field := range rest {
		value, err := light(field)
		if err != nil {
			return at, nil, err
		}
		row = append(row, value)
	}
	return at, row, nil
}

// light reads one amount of light: a finite number that is not negative.
func light(field string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(field), floatBits)
	if err != nil {
		return value, err
	}
	if math.IsNaN(value) || math.IsInf(value, anySign) || value < dark {
		return value, fmt.Errorf("%v is not a finite amount of light",
			value)
	}
	return value, nil
}

// Must reads a table that is embedded in the library. A bad one is a bad
// build. It stops the program at start-up and names the table.
// It never leaves the library with fewer rows than it was given.
func Must(name, csv string, values int) map[int][]float64 {
	parsed, err := parse(csv, values)
	if err != nil {
		panic("table " + name + ": " + err.Error())
	}
	return parsed
}

// The shape of a line. Each is defined by the format, which is the CIE's
// tables as data/README.md in the swatch package describes them.
const (
	// nmStep is the nanometres between rows. It is defined by the tables:
	// both are at every nanometre.
	nmStep = 1
	// lineOne is the number of a file's first line. It is defined by
	// counting lines from one.
	lineOne = 1
	// nothing is no rows. It is defined by counting.
	nothing = 0
	// wavelengthAt is where a row's wavelength is. It is defined by the
	// format: the first field.
	wavelengthAt = 0
	// wavelengthField is how many fields the wavelength takes. It is
	// defined by the format: one.
	wavelengthField = 1
	// floatBits is the size of the numbers read. It is defined by float64.
	floatBits = 64
	// anySign asks math.IsInf for an infinity of either sign. It is
	// defined by math.IsInf.
	anySign = 0
	// dark is the least a table may say. It is defined: light is not
	// negative.
	dark = 0
)
