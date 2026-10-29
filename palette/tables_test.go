package palette_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/palette"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// curves is each transfer table, by name, made at a size.
var curves = map[string]func(palette.Steps) (*palette.Curve, error){
	"to linear":   palette.ToLinear,
	"from linear": palette.FromLinear,
}

// exact is the curve each table stands for.
var exact = map[string]func(float64) float64{
	"to linear":   srgb.ToLinear,
	"from linear": srgb.FromLinear,
}

func TestEveryTableSaysItsErrorInJNDs(test *testing.T) {
	for name, build := range curves {
		small, err := build(16)
		if err != nil {
			test.Fatal(err)
		}
		large, err := build(4096)
		if err != nil {
			test.Fatal(err)
		}
		if small.Worst() <= large.Worst() {
			test.Errorf("%s: 16 steps %v JND, 4096 steps %v: a larger "+
				"table should be closer", name, small.Worst(), large.Worst())
		}
		if large.Worst() >= 0.1 {
			test.Errorf("%s at 4096 steps: %v JND, want under a tenth",
				name, large.Worst())
		}
	}
}

func TestATableIsExactAtItsStepsAndOutsideIt(test *testing.T) {
	for name, build := range curves {
		table, err := build(65)
		if err != nil {
			test.Fatal(err)
		}
		// at 65 steps the table holds k/64, so these are its own points
		for _, level := range []float64{0, 0.25, 0.5, 1, -0.2, 1.7} {
			if got, want := table.At(level), exact[name](level); got != want {
				test.Errorf("%s at %v: %v, want %v exactly", name, level,
					got, want)
			}
		}
	}
}

func TestATooSmallTableIsRefused(test *testing.T) {
	for name, build := range curves {
		if _, err := build(1); err == nil {
			test.Errorf("%s with 1 step: want a refusal", name)
		}
	}
	red, _, _ := srgb.Primaries()
	if _, err := palette.Blend(red, swatch.White, ok.Mix, 1); err == nil {
		test.Errorf("a blend with 1 step: want a refusal")
	}
}

func TestTwoStepsAreEnough(test *testing.T) {
	for name, build := range curves {
		if _, err := build(2); err != nil {
			test.Errorf("%s with 2 steps, one at each end: %v", name, err)
		}
	}
	red, _, _ := srgb.Primaries()
	if _, err := palette.Blend(red, swatch.White, ok.Mix, 2); err != nil {
		test.Errorf("a blend with 2 steps: %v", err)
	}
}

func TestABlendSaysItsErrorAndIsExactAtItsEnds(test *testing.T) {
	red, _, blue := srgb.Primaries()
	blend, err := palette.Blend(red, blue, ok.Mix, 64)
	if err != nil {
		test.Fatal(err)
	}
	if blend.Worst() <= 0 || blend.Worst() >= 1 {
		test.Errorf("red to blue in oklab, 64 steps: %v JND, want "+
			"above 0 and under 1", blend.Worst())
	}
	for _, along := range []float64{0, 1, -0.5, 1.5} {
		got, want := blend.At(along), ok.Mix.Mix(red, blue, along)
		if got != want {
			test.Errorf("at %v: %v, want %v exactly", along, got, want)
		}
	}
	finer, err := palette.Blend(red, blue, ok.Mix, 1024)
	if err != nil {
		test.Fatal(err)
	}
	if finer.Worst() >= blend.Worst() {
		test.Errorf("1024 steps %v JND, 64 steps %v: want closer",
			finer.Worst(), blend.Worst())
	}
}

// finely is a curve table's error measured here, apart from the
// palette: as a grey, at 128 points between each pair of the table's.
func finely(table *palette.Curve, steps palette.Steps,
	curve func(float64) float64, light bool) float64 {
	grey := func(value float64) swatch.Swatch {
		if !light {
			value = srgb.ToLinear(value)
		}
		return srgb.FromLight(value, value, value)
	}
	samples := (int(steps)-1)*128 + 1
	var worst float64
	for index := range samples {
		at := float64(index) / float64(samples-1)
		worst = max(worst, ok.Distance(ok.FromSwatch(grey(table.At(at))),
			ok.FromSwatch(grey(curve(at))))/ok.Eye)
	}
	return worst
}

func TestAnErrorMeasuredFinerAgrees(test *testing.T) {
	for _, steps := range []palette.Steps{16, 64, 256} {
		to, _ := palette.ToLinear(steps)
		from, _ := palette.FromLinear(steps)
		for name, check := range map[string][2]float64{
			"to linear": {float64(to.Worst()),
				finely(to, steps, srgb.ToLinear, true)},
			"from linear": {float64(from.Worst()),
				finely(from, steps, srgb.FromLinear, false)},
		} {
			said, found := check[0], check[1]
			if found < said || found > said*1.02 {
				test.Errorf("%s at %d steps: says %v JND, measured "+
					"eight times finer %v; want within 2%%", name, steps,
					said, found)
			}
		}
	}
}

func TestNaNGoesToTheExactCurve(test *testing.T) {
	// int(math.Floor(NaN)) differs by platform, so NaN must never reach
	// a table's points: on amd64 it read far out of range
	nan := math.NaN()
	for name, build := range curves {
		table, err := build(16)
		if err != nil {
			test.Fatal(err)
		}
		if got := table.At(nan); !math.IsNaN(got) {
			test.Errorf("%s at NaN: %v, want the curve's NaN", name, got)
		}
	}
	red, _, blue := srgb.Primaries()
	ramp, err := palette.Blend(red, blue, ok.Mix, 16)
	if err != nil {
		test.Fatal(err)
	}
	if x, _, _ := ramp.At(nan).XYZ(); !math.IsNaN(x) {
		test.Errorf("a blend at NaN: X %v, want the mixer's NaN", x)
	}
}
