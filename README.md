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

    primitives/oklab      the same colour in the coordinates an eye agrees
                          with: lightness, green-to-red, blue-to-yellow.

    make                  gofmt, go vet, go test

jane michelle arc, 2026.
