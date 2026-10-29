# changelog

## v0.4.5 SAME DAY HOLY BISCUITS

- moved si out of internal into primitives so `nilgai` can use them in
  searching for blazars.
- fixed slightly broken ci 
- created primitives/length so callers can refer to angstrom or pc or
  whatever, normal css stuff
- si.Gravitation is G, kept apart from the defining constants because
  it is measured, with its uncertainty stated.
- radiation's MetresToNm and NmToMetres bring lengths into nanometres:
  MetresToNm(3.4 * si.Micro(length.Metre)) is 3400 nm.

## v0.4.1, 2026-10-09

mistakes were made. 0.4.0 went out without the license we had carefully
selected over several months of anguished discussions. the licence is
apache 2.0. party parrot.

## v0.4.0, 2026-10-09

light:

- primitives/bands holds light as power in slices of the spectrum, 24
  to an octave. visualdocs has a page for it.
- spaces/radiation shows light an eye can't see, through a false colour
  profile.
- swatch can give you d65 and a black body as spectra.
- srgb's transfer curve and linear light are public.
- the si's defining constants live in one place, internal/si.

colourways:

- the colourway package holds a colourway's source, writes it back, and
  renders it into each program's files. paratune saves through it.
- general roles (ground, ink, heading, err and the rest) sit above the
  dialects, and colourway.Resolve fills in the ones a colourway leaves
  out.
- Install copies colourways into your own folder and each render to
  where its program reads it. a file you tuned is kept unless you ask
  to replace it.
- cmd/colourway diff and show compare two colourways as colours. git
  can use show as a textconv.

dialects:

- ghostty, nvim and glamour each write a colourway and read one back.
- dialects/vim is the neovim scheme in vimscript.
- dialects/claude names the palette numbers claude code draws in, and
  they follow the general roles.

the banner in art/ is back. at least one person was overjoyed. that person was
jane michelle.
