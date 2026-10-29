package table_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/internal/table"
)

// The numbers the tests pick. Each is defined by this test and says what for.
const (
	// threeValues is the values in a test row, as in X, Y and Z. It is
	// defined by this test.
	threeValues = 3
	// firstWavelength is the lowest wavelength a drawn table starts at. It
	// is defined by this test.
	firstWavelength = 300
	// lastWavelength is the highest a drawn table starts at. It is defined
	// by this test.
	lastWavelength = 830
	// firstOfGood is the wavelength of the first row of good, below. It is
	// defined by good.
	firstOfGood = 360
	// fewestRows is the fewest rows a drawn table has. It is defined by
	// this test.
	fewestRows = 1
	// mostRows is the most rows a drawn table has. It is defined by this
	// test.
	mostRows = 40
	// leastLight is the least a value may be. It is defined: not negative.
	leastLight = 0
	// shortest asks strconv for the fewest digits that read back exactly.
	// It is defined by strconv.
	shortest = -1
	// floatBits is the size of the numbers written. It is defined by
	// float64.
	floatBits = 64
	// onceOnly is how many commas the spoiling replaces. It is defined by
	// this test.
	onceOnly = 1
	// noRows is a table with nothing in it. It is defined by counting.
	noRows = 0
)

// good is a sound table of two rows of three values.
const good = "360,0.1,0.2,0.3\n361,0.4,0.5,0.6\n"

// reading is a table as text, ready to be read as if embedded in the
// library.
type reading struct{ csv string }

// run reads the table, which stops the program when it is bad.
func (text reading) run() { table.Must("test.csv", text.csv, threeValues) }

// stoppedWith is what the function panics with, or nothing.
func stoppedWith(run func()) (said string) {
	defer record(&said)
	run()
	return said
}

// record is deferred to take a panic's text, if there is one.
func record(said *string) {
	if text, panicked := recover().(string); panicked {
		*said = text
	}
}

// refusal is what reading the text says if it is refused, or nothing.
func refusal(csv string) string { return stoppedWith(reading{csv}.run) }

// Blank lines, carriage returns and spaces do not stop a sound table
// from reading. The first row is there.
func TestASoundTableReads(t *testing.T) {
	for _, csv := range []string{good, "\n" + good + "\n",
		strings.ReplaceAll(good, "\n", "\r\n"), "360, 0.1, 0.2, 0.3\n"} {
		if said := refusal(csv); said != "" {
			t.Errorf("%q was refused: %v", csv, said)
		}
		rows := table.Must("good.csv", csv, threeValues)
		if len(rows[firstOfGood]) != threeValues {
			t.Errorf("%q: the first row is %v", csv, rows)
		}
	}
}

// The anti-tests. Each wrong table stops the program. The message names
// the table and the line; no row is dropped quietly. A row lost here is
// lost from every colour the library converts.
func TestAMalformedTableIsRefused(t *testing.T) {
	for name, csv := range map[string]string{
		"a gap":             "360,0.1,0.2,0.3\n362,0.4,0.5,0.6\n",
		"a repeat":          "360,0.1,0.2,0.3\n360,0.4,0.5,0.6\n",
		"backwards":         "361,0.1,0.2,0.3\n360,0.4,0.5,0.6\n",
		"a short row":       "360,0.1,0.2\n361,0.4,0.5,0.6\n",
		"a long row":        "360,0.1,0.2,0.3,0.9\n",
		"a word":            "360,0.1,x,0.3\n",
		"no wavelength":     "a,0.1,0.2,0.3\n",
		"a fractional nm":   "360.5,0.1,0.2,0.3\n",
		"not a number":      "360,0.1,NaN,0.3\n",
		"infinite":          "360,0.1,Inf,0.3\n",
		"too big":           "360,0.1,1e999,0.3\n",
		"negative":          "360,0.1,-0.2,0.3\n",
		"a header":          "nm,x,y,z\n360,0.1,0.2,0.3\n",
		"an empty value":    "360,0.1,,0.3\n",
		"a lone wavelength": "360\n",
	} {
		said := refusal(csv)
		if !strings.Contains(said, "test.csv") ||
			!strings.Contains(said, "line") {
			t.Errorf("%s was read as a table: %q", name, said)
		}
	}
}

// An empty table is a missing one, so it stops the program too.
func TestNothingIsNotATable(t *testing.T) {
	for _, csv := range []string{"", "\n\n", " \r\n"} {
		if said := refusal(csv); !strings.Contains(said, "no rows") {
			t.Errorf("%q was read as a table: %q", csv, said)
		}
	}
}

// The message gives the number of the wrong line, counting from one.
func TestTheErrorNamesTheLine(t *testing.T) {
	said := refusal("360,0.1,0.2,0.3\n361,0.1,x,0.3\n")
	if !strings.Contains(said, "line 2:") {
		t.Errorf("the error does not name line 2: %q", said)
	}
}

// The property: a table of any length from any start is read when its
// wavelengths rise by one, every row has its values and every value is a
// finite number that is not negative; spoil one value and it is refused.
func TestATableIsReadExactlyWhenItIsSound(t *testing.T) {
	rapid.Check(t, readsWhenSound)
}

// readsWhenSound is one draw of the property above.
func readsWhenSound(t *rapid.T) {
	rows := rapid.IntRange(fewestRows, mostRows).Draw(t, "rows")
	start := rapid.IntRange(firstWavelength, lastWavelength).Draw(t, "start")
	value := rapid.Float64Range(leastLight, math.MaxFloat64)
	var csv strings.Builder
	for row := range rows {
		csv.WriteString(strconv.Itoa(start + row))
		for range threeValues {
			amount := value.Draw(t, "value")
			csv.WriteString("," + strconv.FormatFloat(amount, 'g', shortest,
				floatBits))
		}
		csv.WriteString("\n")
	}
	if said := refusal(csv.String()); said != "" {
		t.Fatalf("a sound table was refused: %q\n%s", said, csv.String())
	}
	read := table.Must("drawn.csv", csv.String(), threeValues)
	if len(read) != rows {
		t.Fatalf("a sound table came back with %d rows, not %d", len(read),
			rows)
	}
	spoiled := strings.Replace(csv.String(), ",", ",x,", onceOnly)
	if refusal(spoiled) == "" {
		t.Fatalf("a spoiled table was read:\n%s", spoiled)
	}
}

// A sound table comes back whole and does not stop the program.
func TestASoundEmbeddedTableComesBack(t *testing.T) {
	if rows := table.Must("good.csv", good, threeValues); len(rows) == noRows {
		t.Errorf("a sound table came back empty: %v", rows)
	}
}
