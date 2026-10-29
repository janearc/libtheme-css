package grafana_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/janearc/libtheme-css/dialects/grafana"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// beyondFile is testdata/beyond.json.
type beyondFile struct {
	Place     swatch.XY
	Luminance float64
}

// beyond is spectral red, a colour no sRGB lamp can make.
func beyond(t *testing.T) swatch.Swatch {
	t.Helper()
	raw, err := os.ReadFile("testdata/beyond.json")
	if err != nil {
		t.Fatal(err)
	}
	var data beyondFile
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	return swatch.FromXY(data.Place, data.Luminance)
}

// Shade returns the bare hue as given, even outside the gamut. Only a
// moved shade is brought inside. Unknown shades are left alone too.
func TestBareHueIsReturnedExactly(t *testing.T) {
	colour := beyond(t)
	if grafana.InGamut(colour) {
		t.Fatal("the test colour is inside the gamut")
	}
	if got := grafana.Shade(colour, ""); got != colour {
		t.Errorf("the bare hue became %v", got)
	}
	if got := grafana.Shade(colour, "nonesuch"); got != colour {
		t.Errorf("an unknown shade made %v", got)
	}
	if got := grafana.Shade(colour, "light"); got == colour {
		t.Errorf("a moved shade is %v", got)
	}
}

// Split gives a name that begins with a dash no shade. It comes back
// whole as the hue.
func TestSplitDoesNotTakeADashForAShade(t *testing.T) {
	shade, hue := grafana.Split("-blue")
	if shade != "" || hue != "-blue" {
		t.Errorf("-blue split as %q and %q", shade, hue)
	}
	shade, hue = grafana.Split("light--blue")
	if shade != "light" || hue != "-blue" {
		t.Errorf("light--blue split as %q and %q", shade, hue)
	}
}
