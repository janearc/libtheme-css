# libtheme
the documentation is a program. `go tool game run visualdocs` pages through it:
the spectrum as the 1931 observer saw it, the grey line, the wheel, two colours
0.02 apart, the same two stops blended in oklab and in the lamps.

`go run ./cmd/visualdocs ramp --css` says a page in css; `--go` prints the
source that painted it, so picture, sheet and code cannot drift apart.

one person keeps a theme in a terminal, an editor, a web page and a lamp, and
matches them by eye, one hex code at a time. this library gives them one
primitive, the swatch, and works up from it, so the same colour is the same
colour everywhere it lands.

    primitives/swatch     one colour, complete, as cie xyz (1931)
    primitives/functions  a ramp: stops and a mixer, t to a swatch
    spaces/ok             oklab and oklch, the axes an eye agrees with
    spaces/srgb           the screen's lamps; hex, hsl, hsv; derived
    css                   the container: a sheet of named swatches
    dialects/             hue, grafana, ghostty, claude, nvim, glamour
    ../cmd/visualdocs     the documentation, shown, one page per idea
    ../cmd/libtheme       the console end: show, ramp, read, known

game builds it; `go tool game run visualdocs` and `run known` read it. every
number here stands on a number somebody picked: FUDGE.md is the ledger,
1931 to your shell. DESIGN.md is the shape and the decisions.
