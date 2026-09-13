# libtheme-css

one person keeps a theme in a terminal, an editor, a web page and a lamp,
and matches them by eye, one hex code at a time. this library gives them
one primitive, the swatch, and translates it out to whatever each device
speaks, so the same colour is the same colour everywhere it lands.

a swatch is one colour, complete. it is stored as cie xyz, the 1931
standard every device space is defined against, which makes this the
world's only pre-war css-normalising theme library. it does its
arithmetic in oklab, where a straight line between two colours looks
straight, and it prints itself the way css says it: `oklch(74% 0.20 345)`.

    s := libtheme.MustHex("#ff6ec7")       // the vaporwave pink
    s.String()                              // oklch(74% 0.20 345)
    libtheme.Mix(s, other, 0.5)             // halfway, by eye, not by lamp
    s.SRGB8()                               // r, g, b, and whether a screen can make it

    make            gofmt, go vet, go test

what is not here, on purpose: how bright a given device's white is, which
colours it cannot make, and what a particular eye can read. those are
facts about the device and the reader, not the colour, and they get their
own files.

jane michelle arc, 2026.
