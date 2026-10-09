# libtheme

one person keeps a theme in a terminal, an editor, a web page and a lamp,
and matches them by eye. libtheme starts from one primitive, the swatch, so
a colour is the same colour wherever it lands.

colour is just light that you can see with your eyes, however they see, and
light is just radiation, and radiation, when you're a programmer, is just
math. so libtheme manages all forms of radiation as easily as it manages srgb.
xrays, microwave, gamma, in your vim theme. yep.

## quick reference

```
# build the package
$ go install github.com/janearc/game/cmd/game@latest
$ ./bin/game build

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
    spaces/ok             oklab and oklch
    spaces/srgb           the screen: hex, hsl, hsv
    spaces/radiation      radiation the eye does not ordinarily see
    css                   a sheet of named swatches
    colourway             one sheet, rendered into every dialect
    dialects/             hue, grafana, ghostty, claude, nvim, vim, glamour
    cmd/visualdocs        the documentation, one page per idea
    cmd/libtheme          show, ramp, read, known
    cmd/colourway         render, sheet, diff, show
