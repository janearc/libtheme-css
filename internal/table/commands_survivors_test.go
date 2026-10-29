package table_test

import (
	"strings"
	"testing"
)

// Every row after the first must rise by one nanometre. This holds at any
// wavelength, however low. The anti-tests are low starts that do rise.
func TestEveryRowAfterTheFirstMustRiseByOne(t *testing.T) {
	for name, csv := range map[string]string{
		"a gap after one":           "1,0.1,0.2,0.3\n5,0.1,0.2,0.3\n",
		"a gap after minus one":     "-1,0.1,0.2,0.3\n5,0.1,0.2,0.3\n",
		"a gap after zero":          "0,0.1,0.2,0.3\n5,0.1,0.2,0.3\n",
		"a repeat of minus one":     "-1,0.1,0.2,0.3\n-1,0.1,0.2,0.3\n",
		"backwards from minus one":  "-1,0.1,0.2,0.3\n-3,0.1,0.2,0.3\n",
		"backwards from a high one": "830,0.1,0.2,0.3\n829,0.1,0.2,0.3\n",
	} {
		said := refusal(csv)
		if !strings.Contains(said, "line 2") {
			t.Errorf("%s: %q was not refused at line 2: %q", name, csv, said)
		}
	}
	for name, csv := range map[string]string{
		"from minus one": "-1,0.1,0.2,0.3\n0,0.1,0.2,0.3\n1,0.1,0.2,0.3\n",
		"from zero":      "0,0.1,0.2,0.3\n1,0.1,0.2,0.3\n",
		"from one":       "1,0.1,0.2,0.3\n2,0.1,0.2,0.3\n",
	} {
		if said := refusal(csv); said != "" {
			t.Errorf("%s was refused: %q", name, said)
		}
	}
}
