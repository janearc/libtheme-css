package readable_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/readable"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// The circle, in degrees.
const (
	// halfCircle is the furthest apart two hues can be. It is defined by
	// the circle.
	halfCircle = 180
	// fullCircle is once round the wheel. It is defined by the circle.
	fullCircle = 360
)

// flagged is whether the colours held to the numbers give the line.
func flagged(t *testing.T, colours map[string]string, numbers rangeNumbers, line string) bool {
	t.Helper()
	report := readable.Check(schemeOf(colours), policyOf(t, numbers))
	return hasLine(report, "test: "+line)
}

// hueOf is the hue of a hex on the lamps, in degrees.
func hueOf(hex string) (hue, saturation float64) {
	scheme := schemeOf(map[string]string{"x": hex})
	lamps, _ := srgb.FromSwatch(scheme.Colours["x"])
	shade := lamps.HSL()
	return shade.H, shade.S
}

// apartBy is the shorter way round the wheel between two hues.
func apartBy(first, second float64) float64 {
	apart := math.Abs(first - second)
	if apart > halfCircle {
		apart = fullCircle - apart
	}
	return apart
}

// Check fails a scheme with missing colours for what is missing and
// nothing more. The other checks do not run on a scheme with holes.
func TestAMissingColourStopsTheOtherChecks(t *testing.T) {
	scheme := schemeOf(inBand)
	delete(scheme.Colours, "panel")
	numbers := edgeData(t).Typical
	numbers.Low = edgeData(t).Unreachable.Low
	numbers.High = edgeData(t).Unreachable.High
	numbers.StateMin = edgeData(t).Unreachable.StateMin
	report := readable.Check(scheme, policyOf(t, numbers))
	if !hasLine(report, "panel: missing colour") {
		t.Error("the missing panel was not named")
	}
	for _, line := range report.Lines() {
		if !strings.Contains(line, "missing colour") {
			t.Errorf("a check ran on a scheme with holes: %q", line)
		}
	}
}

// Check keeps a colour inside when it sits exactly at the start or end
// of the band or at the floor. One step past any of them is out.
func TestTheBandAndTheFloorAreInclusive(t *testing.T) {
	data := edgeData(t)
	scheme := schemeOf(inBand)
	prose, state := ratio(scheme, "ink", "bg"), ratio(scheme, "err", "bg")
	dim, glare := "ink on the ground: too dim to read", "ink on the ground: glare"
	scan := "err on the ground: too dim to scan"
	over, under := math.MaxFloat64, -math.MaxFloat64

	numbers := data.Typical
	numbers.Low, numbers.High = prose, prose+data.Room
	if flagged(t, inBand, numbers, dim) {
		t.Error("ink exactly at the start of the band was too dim")
	}
	numbers.Low = math.Nextafter(prose, over)
	if !flagged(t, inBand, numbers, dim) {
		t.Error("ink just under the start of the band was not too dim")
	}
	numbers = data.Typical
	numbers.High = prose
	if flagged(t, inBand, numbers, glare) {
		t.Error("ink exactly at the end of the band glared")
	}
	numbers.High = math.Nextafter(prose, under)
	if !flagged(t, inBand, numbers, glare) {
		t.Error("ink just over the end of the band did not glare")
	}
	numbers = data.Typical
	numbers.StateMin, numbers.Low, numbers.High = state, state+data.Room, state+data.Room+data.Room
	if flagged(t, inBand, numbers, scan) {
		t.Error("a state colour exactly at the floor was too dim")
	}
	numbers.StateMin = math.Nextafter(state, over)
	if !flagged(t, inBand, numbers, scan) {
		t.Error("a state colour just under the floor was not too dim")
	}
}

// The hue rule is inclusive at both edges. Ground exactly as saturated
// as the neutral line is not neutral. Ink exactly the least gap round
// the wheel from it has the gap. One step the other way and the rule
// bites.
func TestTheHueRuleIsInclusiveAtItsEdges(t *testing.T) {
	data := edgeData(t)
	colours := changed(map[string]string{"bg": data.EdgeGround})
	groundHue, saturation := hueOf(data.EdgeGround)
	inkHue, _ := hueOf(colours["ink"])
	apart := apartBy(groundHue, inkHue)
	line := "ink on the ground: no colour edge"
	over := math.MaxFloat64

	numbers := data.Typical
	numbers.Neutral, numbers.Gap = saturation, apart+data.Room
	if !flagged(t, colours, numbers, line) {
		t.Error("a ground exactly at the neutral line was taken as neutral")
	}
	numbers.Neutral = math.Nextafter(saturation, over)
	if flagged(t, colours, numbers, line) {
		t.Error("a ground just under the neutral line had a colour edge")
	}
	numbers = data.Typical
	numbers.Gap = apart
	if flagged(t, colours, numbers, line) {
		t.Error("ink exactly the least gap away had no colour edge")
	}
	numbers.Gap = math.Nextafter(apart, over)
	if !flagged(t, colours, numbers, line) {
		t.Error("ink a hair short of the gap had a colour edge")
	}
}

// The gap is the short way round the wheel. That holds across the top
// and plain. The line says it in whole degrees.
func TestTheHueGapIsTheShortWayRound(t *testing.T) {
	data := edgeData(t)
	for _, pair := range data.HuePairs {
		colours := changed(map[string]string{"bg": pair.Ground, "ink": pair.Ink})
		groundHue, _ := hueOf(pair.Ground)
		inkHue, _ := hueOf(pair.Ink)
		want := fmt.Sprintf("ink on the ground: no colour edge (%.0f degrees", apartBy(groundHue, inkHue))
		if !flagged(t, colours, data.Typical, want) {
			t.Errorf("%s: no %q", pair.Name, want)
		}
	}
}

// Check passes a light scheme that is in range everywhere. Its prose is
// dark on light, which a pair judged against black would call too dim.
func TestALightSchemeInRangePasses(t *testing.T) {
	report := checked(t, edgeData(t).Light)
	if !report.OK() {
		t.Errorf("a light scheme in range failed:\n%s", joined(report))
	}
}

// Check tells a hue failure of prose on a surface once, by the surface
// check. The pair checks leave it alone.
func TestAProseFailureIsToldOnce(t *testing.T) {
	colours := changed(map[string]string{"bg": "#14507a", "heading": "#2f7ac0"})
	report := checked(t, colours)
	if !hasLine(report, "heading on the ground: no colour edge") {
		t.Fatal("the failure was not found")
	}
	if hasLine(report, "draws heading on bg") {
		t.Errorf("the pair check told it again:\n%s", joined(report))
	}
}
