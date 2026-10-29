package palette_test

import (
	"fmt"

	"github.com/janearc/libtheme-css/palette"
	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// A render makes one palette for its scene, gives it the lamps and the
// materials, and asks for each pair as rays meet them. The first ask
// works the pair out; every later one is kept.
func ExampleNew() {
	scene := palette.New()
	sun := scene.Lamp(bands.Split(swatch.D65()))
	grey, err := bands.Even(0.5)
	if err != nil {
		panic(err)
	}
	wall := scene.Material(grey)
	for range 3 {
		if _, err := scene.Swatch(sun, wall); err != nil {
			panic(err)
		}
	}
	seen, _ := scene.Swatch(sun, wall)
	_, luminance, _ := seen.XYZ()
	fmt.Printf("worked out %d times; Y %.3f\n", scene.Worked(), luminance)
	// Output: worked out 1 times; Y 0.500
}

// Each table measures its own worst error when it is made, in JNDs. The
// caller picks the size from what that error is worth to it.
func ExampleToLinear() {
	for _, steps := range []palette.Steps{16, 64, 256, 1024, 4096} {
		to, _ := palette.ToLinear(steps)
		from, _ := palette.FromLinear(steps)
		fmt.Printf("%d steps: to linear %.6f JND, from linear %.6f JND\n",
			steps, to.Worst(), from.Worst())
	}
	// Output:
	// 16 steps: to linear 0.278371 JND, from linear 3.603529 JND
	// 64 steps: to linear 0.023671 JND, from linear 1.180771 JND
	// 256 steps: to linear 0.002018 JND, from linear 0.102652 JND
	// 1024 steps: to linear 0.000296 JND, from linear 0.014901 JND
	// 4096 steps: to linear 0.000059 JND, from linear 0.000993 JND
}

// A blend between two fixed colours, as a ramp does it, made a table.
func ExampleBlend() {
	red, _, blue := srgb.Primaries()
	for _, steps := range []palette.Steps{16, 64, 256} {
		ramp, _ := palette.Blend(red, blue, ok.Mix, steps)
		fmt.Printf("%d steps: red to blue in oklab, %.6f JND\n", steps,
			ramp.Worst())
	}
	// Output:
	// 16 steps: red to blue in oklab, 0.041527 JND
	// 64 steps: red to blue in oklab, 0.002407 JND
	// 256 steps: red to blue in oklab, 0.000148 JND
}
