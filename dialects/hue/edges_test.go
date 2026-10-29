package hue_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/janearc/libtheme-css/dialects/hue"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// edgeFile is testdata/edges.json.
type edgeFile struct {
	Places       []swatch.XY
	UnitLum      float64 `json:"unit luminance"`
	SmallestMrk  int     `json:"smallest mirek"`
	KelvinPerMeg float64 `json:"kelvin per megakelvin"`
	Gamuts       map[string]swatch.Gamut
}

// edgeData reads testdata/edges.json, or the test stops.
func edgeData(t *testing.T) edgeFile {
	t.Helper()
	raw, err := os.ReadFile("testdata/edges.json")
	if err != nil {
		t.Fatal(err)
	}
	var data edgeFile
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

// Swatch puts a place on the horseshoe at unit luminance: its Y is one.
func TestSwatchIsAtUnitLuminance(t *testing.T) {
	data := edgeData(t)
	for _, place := range data.Places {
		if _, y, _ := hue.Swatch(place).XYZ(); y != data.UnitLum {
			t.Errorf("the place %v has luminance %v", place, y)
		}
	}
}

// The anti-test beside the mirek of nothing. The smallest mirek is the
// smallest real temperature, a black body at a million kelvin.
func TestTheSmallestMirekIsLight(t *testing.T) {
	data := edgeData(t)
	got := hue.Mirek(data.SmallestMrk)
	if got == swatch.NoLight() {
		t.Fatal("the smallest mirek was no light")
	}
	if want := swatch.Planckian(data.KelvinPerMeg / float64(data.SmallestMrk)); got != want {
		t.Errorf("the smallest mirek is %v, want %v", got, want)
	}
}

// The corners are as philips publishes them. Each is pinned, since one
// mistyped digit moves a corner.
func TestGamutCornersAreAsPublished(t *testing.T) {
	want := edgeData(t).Gamuts
	got := map[string]swatch.Gamut{"A": hue.GamutA, "B": hue.GamutB, "C": hue.GamutC}
	for name, gamut := range want {
		if got[name] != gamut {
			t.Errorf("gamut %s is %v, want %v", name, got[name], gamut)
		}
	}
}
