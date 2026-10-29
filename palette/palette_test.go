package palette_test

import (
	"math"
	"sync"
	"testing"

	"github.com/janearc/libtheme-css/palette"
	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// paint is a made-up red surface: little back below 580 nm, most above.
func paint(test *testing.T) bands.Share {
	test.Helper()
	share, err := bands.Shares(map[int]float64{
		400: 0.05, 500: 0.05, 580: 0.3, 620: 0.8, 700: 0.85})
	if err != nil {
		test.Fatal(err)
	}
	return share
}

// sun is D65 in bands.
func sun() bands.Light { return bands.Split(swatch.D65()) }

// sodium is a low-pressure sodium lamp: its two lines, nothing else.
func sodium() bands.Light {
	return bands.Lines(map[float64]float64{589.0: 1, 589.6: 1})
}

// sameBits is whether two swatches are the same bytes, not merely close.
func sameBits(kept, fresh swatch.Swatch) bool {
	kx, ky, kz := kept.XYZ()
	fx, fy, fz := fresh.XYZ()
	return math.Float64bits(kx) == math.Float64bits(fx) &&
		math.Float64bits(ky) == math.Float64bits(fy) &&
		math.Float64bits(kz) == math.Float64bits(fz)
}

// sameLight is whether two lights are the same bytes in every band.
func sameLight(kept, fresh bands.Light) bool {
	kfirst, klast := kept.Span()
	ffirst, flast := fresh.Span()
	if kfirst != ffirst || klast != flast {
		return false
	}
	for band := kfirst; band <= klast; band++ {
		if math.Float64bits(kept.Power(band)) !=
			math.Float64bits(fresh.Power(band)) {
			return false
		}
	}
	return true
}

func TestAPalettesAnswerIsTheSameBytesAsWorkingItOut(test *testing.T) {
	scene := palette.New()
	red := scene.Material(paint(test))
	for _, light := range []bands.Light{sun(), sodium()} {
		lamp := scene.Lamp(light)
		fresh := bands.Lit(light, paint(test))
		for range 2 { // the first works it out; the second is kept
			lit, err := scene.Light(lamp, red)
			if err != nil {
				test.Fatal(err)
			}
			seen, err := scene.Swatch(lamp, red)
			if err != nil {
				test.Fatal(err)
			}
			if !sameLight(lit, fresh) || !sameBits(seen, fresh.Swatch()) {
				test.Fatalf("the palette's answer differs from fresh")
			}
		}
	}
}

func TestEachPairIsWorkedOutOnce(test *testing.T) {
	scene := palette.New()
	lamp, red := scene.Lamp(sun()), scene.Material(paint(test))
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, err := scene.Swatch(lamp, red); err != nil {
				test.Error(err)
			}
		}()
	}
	wait.Wait()
	if worked := scene.Worked(); worked != 1 {
		test.Errorf("32 renders asked for one pair: worked out %d times, "+
			"want 1", worked)
	}
	other := scene.Lamp(sodium())
	if _, err := scene.Swatch(other, red); err != nil {
		test.Fatal(err)
	}
	if worked := scene.Worked(); worked != 2 {
		test.Errorf("a second pair: worked out %d in all, want 2", worked)
	}
}

func TestAHandleFromAnotherPaletteIsRefused(test *testing.T) {
	scene, elsewhere := palette.New(), palette.New()
	lamp := elsewhere.Lamp(sun())
	red := scene.Material(paint(test))
	if _, err := scene.Swatch(lamp, red); err == nil {
		test.Errorf("a lamp from another palette was answered")
	}
	var nothing palette.Lamp
	if _, err := scene.Light(nothing, red); err == nil {
		test.Errorf("a lamp from no palette was answered")
	}
}

func TestNoPaletteIsRefusedNotAPanic(test *testing.T) {
	var none *palette.Palette
	var lamp palette.Lamp
	var material palette.Material
	if _, err := none.Swatch(lamp, material); err == nil {
		test.Errorf("no palette, no handles: answered")
	}
	if _, err := none.Light(lamp, material); err == nil {
		test.Errorf("no palette, no handles: answered")
	}
}

// BenchmarkAKeptPairInParallel is the property the palette exists for:
// a kept pair read by every core at once, which must not wait on a lock.
// run it with -cpu 1,4,8,16; the time per ask should fall, not rise.
func BenchmarkAKeptPairInParallel(bench *testing.B) {
	scene := palette.New()
	lamp := scene.Lamp(sun())
	share, err := bands.Even(0.5)
	if err != nil {
		bench.Fatal(err)
	}
	wall := scene.Material(share)
	if _, err := scene.Swatch(lamp, wall); err != nil {
		bench.Fatal(err)
	}
	bench.RunParallel(func(each *testing.PB) {
		for each.Next() {
			if _, err := scene.Swatch(lamp, wall); err != nil {
				bench.Error(err)
			}
		}
	})
}
