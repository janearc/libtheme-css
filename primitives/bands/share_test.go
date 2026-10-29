package bands_test

import (
	"math"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// daylight is D65 in bands.
func daylight() bands.Light { return bands.Split(swatch.D65()) }

// even is a share that is part everywhere, or the test stops.
func even(test *testing.T, part float64) bands.Share {
	test.Helper()
	share, err := bands.Even(part)
	if err != nil {
		test.Fatal(err)
	}
	return share
}

// visible is the bands an eye sees, about.
func visible() (first, last bands.Band) {
	return bands.Of(380), bands.Of(780)
}

func TestASurfaceThatPassesEverythingLeavesTheLightAsItIs(test *testing.T) {
	light := daylight()
	lit := bands.Lit(light, even(test, 1))
	first, last := light.Span()
	for band := first; band <= last; band++ {
		if lit.Power(band) != light.Power(band) {
			test.Fatalf("band %d: %v lit, %v before", band,
				lit.Power(band), light.Power(band))
		}
	}
}

func TestHalfTheShareIsHalfTheLight(test *testing.T) {
	light := daylight()
	lit := bands.Lit(light, even(test, 0.5))
	first, last := visible()
	for band := first; band <= last; band++ {
		if lit.Power(band) != 0.5*light.Power(band) {
			test.Fatalf("band %d: %v, want half of %v", band,
				lit.Power(band), light.Power(band))
		}
	}
}

func TestAShareIsAveragedInABandNotSummed(test *testing.T) {
	// 600 and 601 nm fall in one band; light would add them, a share
	// is the fraction passed there, so it is their mean
	band := bands.Of(600)
	if bands.Of(601) != band {
		test.Fatal("600 and 601 nm are expected to share a band")
	}
	share, err := bands.Shares(map[int]float64{600: 0.2, 601: 0.4})
	if err != nil {
		test.Fatal(err)
	}
	if got := share.Part(band); !(math.Abs(got-0.3) <= 1e-12) {
		test.Errorf("band %d: part %v, want the mean 0.3", band, got)
	}
}

func TestABandBetweenSamplesIsInterpolated(test *testing.T) {
	// ends other than 0 and 1, so a sum in place of a difference shows
	share, err := bands.Shares(map[int]float64{400: 0.2, 700: 0.8})
	if err != nil {
		test.Fatal(err)
	}
	middle := bands.Of(550)
	lo, hi := middle.Edges()
	centre := math.Sqrt(lo * hi)
	want := 0.2 + 0.6*(centre-400)/300
	// written so a NaN fails it, not passes
	if got := share.Part(middle); !(math.Abs(got-want) <= 1e-12) {
		test.Errorf("band %d: part %v, want %v between the two",
			middle, got, want)
	}
}

func TestOutsideAShareNothingPasses(test *testing.T) {
	share, err := bands.Shares(map[int]float64{600: 1, 650: 1})
	if err != nil {
		test.Fatal(err)
	}
	lit := bands.Lit(daylight(), share)
	if got := lit.Power(bands.Of(450)); got != 0 {
		test.Errorf("450 nm, outside the share: %v passed, want 0", got)
	}
	if got := lit.Power(bands.Of(620)); got == 0 {
		test.Errorf("620 nm, inside the share: nothing passed")
	}
}

func TestNoLightOrNoShareIsNoLight(test *testing.T) {
	var none bands.Light
	if first, last := bands.Lit(none, even(test, 1)).Span(); first <= last {
		test.Errorf("no light, lit: span %d to %d, want empty", first, last)
	}
	var nothing bands.Share
	lit := bands.Lit(daylight(), nothing)
	if y := lit.Swatch(); y != swatch.Black {
		test.Errorf("light through no share: %v, want black", y)
	}
}

func TestASharePastZeroOrOneIsRefused(test *testing.T) {
	for _, part := range []float64{-0.1, 1.1, math.NaN()} {
		if _, err := bands.Even(part); err == nil ||
			!strings.Contains(err.Error(), "0 and 1") {
			test.Errorf("Even(%v): %v, want a refusal naming 0 and 1",
				part, err)
		}
		if _, err := bands.Shares(map[int]float64{500: part}); err == nil {
			test.Errorf("Shares with %v: want a refusal", part)
		}
	}
	if _, err := bands.Shares(nil); err == nil {
		test.Errorf("Shares of nothing: want a refusal")
	}
	for _, nm := range []int{0, -5} {
		if _, err := bands.Shares(map[int]float64{nm: 0.5, 500: 0.5}); err == nil {
			test.Errorf("a share at %d nm: want a refusal", nm)
		}
	}
	// the shortest wavelength bands carries is a nanometre, band 0
	if _, err := bands.Shares(map[int]float64{1: 0.5, 500: 0.5}); err != nil {
		test.Errorf("a share at 1 nm: %v, want it allowed", err)
	}
}

func TestAShareOfNothingIsAllowed(test *testing.T) {
	// a surface that sends nothing back is black, and still a surface
	lit := bands.Lit(daylight(), even(test, 0))
	if y := lit.Swatch(); y != swatch.Black {
		test.Errorf("daylight on a share of 0: %v, want black", y)
	}
	if _, err := bands.Shares(map[int]float64{500: 0}); err != nil {
		test.Errorf("a share of 0 at 500 nm: %v, want it allowed", err)
	}
}
