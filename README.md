# libtheme-css

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

    cmd/libtheme          the console end: show a colour in every form,
                          painted, or draw the line between two colours
                          in oklab and in the lamps side by side.

    make                  gofmt, go vet, go test
    make build            bin/libtheme

every number here stands on a number somebody picked. FUDGE.md is the
ledger, 1931 to your shell. DESIGN.md is the shape and the decisions,
a screen and a half; `go doc` prints the rest.

jane michelle arc, 2026.
