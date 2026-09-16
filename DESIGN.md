# design

what this library is for, the decisions that shaped it, and how to read
it. a screen and a half. the ledger of where the numbers come from is
FUDGE.md; the first screen is README.md; the rest is in the package
comments, which `go doc` prints.

## the problem

one person keeps a theme in a terminal, an editor, a web page and a lamp, and
matches them by eye, one hex code at a time. a hex code is three lamp levels for
one kind of screen; it cannot say what a lamp should do, or whether the same
colour is the same colour somewhere else.

the library gives that person one primitive that can, and works up from it.

## the shape

five layers, each importing only what is below it:

    primitives/swatch     one colour, complete. stored as cie xyz. knows
                          nothing about devices, eyes or arithmetic.
    primitives/functions  things that turn a number into a swatch: the
                          ramp. imports only the swatch.
    spaces/               coordinate systems on the swatch: ok (oklab and
                          its polar form), srgb (lamps, hex, hsl, hsv).
                          each supplies a mixer for the ramp.
    dialects/             the vendors: how somebody else names a colour,
                          and how that name becomes a swatch and comes
                          back. hue (a lamp says a place and a mirek, not
                          a colour), grafana (five shades of a hue, since
                          it takes no theme and must be told in hex).
    css/                  the container everything normalises to: a sheet
                          of named swatches, written as custom properties.

the console tool under cmd/ is a check, not a product: it shows a colour
in every form and paints it, so a change lower down is a change you can
see. `make visualtest` and `make visualtest-css` run it.

## decisions, and when

- store xyz, work in oklab (13 sep). xyz is the 1931 root every device
  space is defined against, so the working space can change without
  touching a stored value; oklab is where distance matches an eye, so
  every operation goes through it.
- the swatch's fields are unexported so nobody averages xyz by accident.
- derive, don't type (13 sep). the white is integrated from the observer
  and the daylight tables; srgb's matrix is derived from the standard's
  four chromaticities and that white; matrix inverses are computed.
- the only typed numbers are the ones with no derivation: the two cie
  tables, srgb's four pairs and its curve, and oklab's fit.
- precision set once (13 sep). `ok.Distance`, `ok.Exact` for arithmetic
  noise, `ok.Eye` for the smallest difference a person notices. every
  "are these the same" in the library compares against one of the two.
- a fit is checked, not explained (13 sep). oklab's matrices are one
  engineer's fit with his reference values as the test; the comment says
  so and sits at the bottom of the file, since nobody needs it to use
  the package.
- the mixer is a parameter (13 sep). spaces disagree about what a
  straight line is, and css lets a file say `in oklab` or `in srgb`, so
  a ramp is told how to blend rather than deciding.
- containers, not form (13 sep). fields, shapes and animation were built
  here and taken out the same afternoon: a ramp answers "what colour is
  t" and never asks where t came from. geometry belongs to whatever
  draws. the seam is one method wide: give a t, get a swatch.
- hue is `none` below the eye's tolerance (13 sep). a hue on a grey is
  arithmetic on noise; css has the word for it and the library uses it.

## what it is not

not a css engine: it reads custom properties whose values are colours and
refuses everything else out loud. not a renderer: it never draws a shape.

not a device: how bright a lamp's white is, which colours it cannot make, and
what a particular eye can read are facts about the device and the reader, and
get their own libraries (the vendor layer, and libreadme).

## how to read it

swatch.go, then observer.go for where the white comes from; then ok.go top to
bottom, stopping before "the numbers"; then srgb.go's matrix derivation; then
ramp.go's package comment for the seam; then css.go. the tests are the same
conversation with numbers in it.

FUDGE.md last, so you know what all of it is standing on.

jane michelle arc, 2026.
