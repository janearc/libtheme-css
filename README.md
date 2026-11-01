# libtheme

libtheme is a physics based package to coordinate theme across many different
spaces or "dialects." internally it uses css, but it is not any kind of css
processing library.

color is just light. light is just radiation. your eyes are just radiation
detecting organs. and all of this is just math. that's the pitch.

## quick reference

```
# build the package
$ go install github.com/janearc/game/cmd/game@latest
$ game build

# demonstrates what theme understands from derivation
$ game run known

# the programming guide to libtheme
$ game run visualdocs

# the same thing, but in css
$ game run visualdocs-css

# the same thing, but in golang!
$ game run visualdocs-go
```

## less quick reference

the documentation is a program. `game run visualdocs` takes you through it,
from the 1931 observer's spectrum to two stops blended in oklab and on the
screen. `--css` prints a page as css, and `--go` the code that painted it.

    primitives/swatch     one colour, as cie xyz (1931)
    primitives/functions  ramps: stops, a mixer, t from 0 to 1
    primitives/bands      light as amounts in slices of the spectrum
    palette               colour work a render keeps, opt-in
    spaces/ok             oklab and oklch
    spaces/srgb           the screen: hex, hsl, hsv
    css                   a sheet of named swatches
    colourway             one sheet, rendered into every dialect
    dialects/             hue, grafana, ghostty, claude, nvim, vim, glamour
    cmd/visualdocs        the documentation, one page per idea
    cmd/libtheme          show, ramp, read, known
    cmd/colourway         render, sheet, diff, show

## palette

for a render, not a sheet. once a tracer carries spectra, colour is its
hot path, and the costly parts repeat: the same paint under the same
lamp, the same curve on every pixel. a palette keeps that work for one
scene. a css user makes none and pays nothing.

    scene := palette.New()                  // one scene, never a global
    sun := scene.Lamp(bands.Split(swatch.D65()))
    paint, err := bands.Shares(reflectance) // part passed, by nm
    if err != nil {
        return err                          // a part past 0 to 1, say
    }
    red := scene.Material(paint)
    seen, err := scene.Swatch(sun, red)     // worked out once, then kept

the pair's answer is exact: the same bytes as `bands.Lit` and `Swatch`
worked out fresh. a `bands.Share` is a fraction passed per band, its own
type, so light and a share of it cannot be passed as each other.

tables are approximate, and each says its error in JNDs (`ok.Eye`, 0.02,
is one): the worst found at sixteen points between each of its own,
which a test eight times finer agrees with to 2%. the caller picks the
size; outside 0 to 1 a table asks the curve itself.

    steps   to linear   from linear   red to blue, oklab
       16       0.278         3.604                0.042
       64       0.024         1.181                0.002
      256       0.002         0.103                0.000
     4096       0.000         0.001

from linear is steep near black, so it wants about four times the steps
of the others for the same error. the numbers are the package's examples,
rounded; `go test` holds the examples to six places.
