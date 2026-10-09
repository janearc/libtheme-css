# libtheme

libtheme stores a colour once and converts it for whatever shows it: a
terminal, an editor, a web page, a lamp. a theme is a sheet of named
colours. libtheme reads a sheet, converts and blends its colours, and
writes it back out. it handles colour and nothing else.

## requirements

- a colour looks the same in a terminal, an editor, a web page and a
  lamp, and nobody has to match hex codes by eye to get there.
- a colour is stored once, as cie xyz. every device space is defined
  from xyz, so adding a device never changes a stored colour.
- distance and blending are done in oklab, where equal steps look equal.
  in the screen's own rgb they do not.
- every number comes from a published table or standard. FUDGE.md lists
  the few that had to be picked by hand.
- css was chosen as an intermediary format because it's common and seems
  like it could do it, whether it was the "right" format or not. `theme`
  is N O T a css parser or library or any ofthat.
- the documentation is a program. each page draws its examples with the
  library, so the picture, the sheet and the code can't drift apart.

## the packages

    primitives/swatch      one colour, stored as cie xyz (1931). the
                           observer and d65 tables it comes from are in
                           primitives/swatch/data
    primitives/functions   ramps: a list of stops and a mixer. give it a
                           t from 0 to 1 and it returns a swatch
    primitives/bands       light as power in slices of the spectrum, 24
                           slices to an octave. the observer turns it
                           into a swatch
    spaces/ok              oklab and oklch: distance, the tolerance
                           check, fitting a colour into a gamut, and a
                           grid of how far a gamut reaches
    spaces/srgb            the screen: hex, hsl, hsv and bytes
    spaces/radiation       radiation an eye can't see, shown through a
                           false colour profile
    css                    a sheet of named swatches, read and written
                           as css custom properties
    colourway              a colourway: its source, written back,
                           resolved, and rendered into each program's
                           files
    dialects/hue           how a hue lamp names a colour, both ways
    dialects/grafana       five shades of one hue, as hex, for grafana
    dialects/ghostty       a ghostty theme: its settings, the sixteen
                           terminal colours, and palette-N lines
    dialects/claude        the palette numbers claude code draws in, by
                           name
    dialects/nvim          a neovim colour scheme: its lua table, the
                           fixed groups, and the sixteen
    dialects/vim           the neovim scheme in vimscript, for vim,
                           without the tree-sitter groups vim can't hold
    dialects/glamour       a glamour markdown style and its
                           highlighter's token colours
    cmd/libtheme           libtheme from a shell: show, ramp, read, known
    cmd/colourway          colourways from a shell: render, sheet, diff,
                           and show, which git uses as a textconv
    cmd/visualdocs         the documentation, one page per idea
    internal/mat           the matrix arithmetic the spaces share
    internal/si            the si's defining constants, typed once

## what imports what

a package imports only packages from an earlier step:

1. internal/si and internal/mat import nothing.
2. swatch imports si. functions and bands import swatch.
3. the spaces import swatch, plus functions or bands.
4. css imports the spaces.
5. most dialects import css. hue and grafana work from the spaces.
6. colourway imports the dialects, and each command imports what it
   shows.

two dialects import other dialects. nvim takes the sixteen terminal
colours from ghostty, because a shell running inside neovim draws with
the terminal's colours. vim takes the whole scheme from nvim.

bands' tests import spaces/ok to measure their results. the bands
package itself does not.

## colourway roles

every dialect names its colours its own way, and the names barely
overlap. so the colourway package has general names, which every
dialect maps from:

- grounds: ground, surface-1, panel, cursor-line
- text: ink, dim, heading, link, code, string, gutter, border
- states: err, warn, info, hint, ok

a program's own name, like nvim-ink, overrides a general role for that
program only.

colourway.Resolve fills in the roles a colourway leaves out. for each
one, it tries these in order:

1. another general role, or one of the sixteen: heading from magenta,
   link from blue, code from green, and so on.
2. for a ground, the ground mixed toward the ink in oklab.
3. the ink for text, and the ground for a ground.

Resolved.From records where each filled-in role came from.

Resolve then fills in claude code's colours from the general roles,
using the claude dialect's table: its text from ink, a code span from
code, the mode line from info, and a diff's bands from the ground mixed
toward ok and err.

every program renders from the resolved roles. so every terminal
colourway also gets a neovim scheme and a markdown style, and the
terminal's selection colour follows surface-1 when the source doesn't
set it.

a ghostty theme gets palette lines for the numbers claude code draws
in: 252, 188, 153, 180, 175, 219, 22, 52 and 231. any program that
draws in those xterm numbers gets the colourway's colours.

## decisions

- store xyz, work in oklab. xyz is what every device space is defined
  from, and oklab is where a distance looks like a distance.
- a swatch's fields are unexported, so nothing can average two xyz
  values by accident.
- derive numbers instead of typing them. the white point is integrated
  from the cie tables, and srgb's matrix is computed from the
  standard's chromaticities and that white.
- the only numbers typed in by hand are the two cie tables, srgb's four
  chromaticity pairs and its transfer curve, oklab's fitted matrices,
  the bands' step, and the si's defining constants.
- light is held in bands spaced evenly in the logarithm of wavelength,
  with one step for everything. a redshift is then a slide along the
  bands, and any two lights line up band for band. the step was picked
  by hand, and FUDGE.md says what it was picked from.
- there are two tolerances, one for arithmetic noise and one for what an
  eye can see. every equality check in the library uses one of them.
- a ramp takes its mixer as a parameter. spaces disagree about what a
  straight line between two colours is, and css lets a sheet say which
  space to blend in.
- libtheme has no fields, shapes or animation. a ramp tells you what
  colour is at t, and never asks where t came from.
- a grey has no hue. libtheme writes its hue as none, the way css does.
- the acceptance tests are the documentation and the known set:
  `go run ./cmd/visualdocs` and `go run ./cmd/libtheme known`.
  `make visualdocs` and `make visualtest` run the same two.

## what it is not

- not a css engine. it reads custom properties whose values are
  colours, and anything else is an error.
- not a renderer. the library draws nothing. the commands in cmd/ do
  the drawing.
- not a device. a lamp's brightness, a screen's gamut and a reader's
  eyes are facts about the device and the reader. they belong in the
  dialects and in libreadme.

## see also

README.md is the first screen. FUDGE.md lists where every number comes
from. the package comments are the rest: `go doc ./primitives/swatch`,
for one.

jane michelle arc, 2026.
