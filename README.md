# libtheme-css

the documentation is a program. `make visualdocs` pages through it: the
spectrum as the 1931 observer saw it, the grey line, the wheel, two
colours 0.02 apart, the same two stops blended in oklab and through the
lamps. `./bin/visualdocs ramp --css` says a page in css;
`./bin/visualdocs ramp --go` prints the source that painted it.

one person keeps a theme in a terminal, an editor, a web page and a lamp,
and matches them by eye, one hex code at a time. this library gives them
one primitive, the swatch, and works up from it, so the same colour is
the same colour everywhere it lands.

it is being written from the bottom, one primitive at a time, on
branches, and read as it goes. what exists:

    primitives/swatch     one colour, complete, stored as cie xyz (1931),
                          which makes this the world's only pre-war
                          css-normalising theme library.

    primitives/functions  a ramp: stops and a mixer, a function from t to a
                          swatch. the spaces each supply a mixer.

    spaces/ok             the same colour on the axes an eye agrees with:
                          oklab, and its polar form oklch.
    spaces/srgb           the same colour as a screen's three lamps, and
                          the hex, hsl and hsv forms of that. its matrix
                          is derived from the standard's four chromaticities
                          and the swatch's white.

    css                   the container everything normalises to: a sheet
                          of named swatches, written out as custom
                          properties with the oklch beside each.

    cmd/visualdocs        the documentation, shown. one page per idea, a
                          few painted rows and a few lines of text, because
                          the honest way to explain a colour is to put it on
                          the screen. every page can also say the same thing
                          as css, and show the go that painted it, which is
                          its own source file, embedded, so picture, sheet
                          and code cannot drift apart.

    cmd/libtheme          the console end: show a colour in every form,
                          painted, or draw the line between two colours
                          in oklab and in the lamps side by side.

    make                  gofmt, go vet, go test
    make build            bin/libtheme, bin/visualdocs
    make visualdocs       read the docs: enter for the next page, q to stop
    make visualtest       everything the library can derive, painted

every number here stands on a number somebody picked. FUDGE.md is the
ledger, 1931 to your shell. DESIGN.md is the shape and the decisions,
a screen and a half; `go doc` prints the rest.

jane michelle arc, 2026.
