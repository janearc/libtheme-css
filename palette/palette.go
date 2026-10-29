// Package palette keeps the colour work a render repeats, so it is done
// once: an opt-in value one render makes for one scene and passes along.
// A caller that makes no palette pays nothing, and nothing elsewhere in
// libtheme changes for one.
//
// It is never package state. A function takes the palette it reads, so
// two renders never share one by accident, and dropping it drops what it
// kept.
//
// It keeps two kinds of thing. Exact: each lamp's light through each
// material's share, and its swatch, the same bytes as working it out
// fresh. Approximate: tables for curves every pixel passes through, each
// saying its own worst error in just-noticeable differences.
package palette

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// Palette is one scene's kept colour work. It is safe for many renders of
// the scene to ask at once, and works each pair out once among them.
//
// A kept pair is read without a lock, since colour is a render's hot path
// and a lock there would be every core's ceiling. Only giving it a lamp or
// a material, and working a pair out the first time, takes one.
type Palette struct {
	lock      sync.Mutex // the lamps and the materials
	lamps     []bands.Light
	materials []bands.Share
	pairs     sync.Map // a pair, to its *worked
	count     atomic.Int64
}

// Lamp is a light given to a palette, as the palette knows it.
type Lamp struct {
	from  *Palette
	index int
}

// Material is a share given to a palette: a surface or a filter.
type Material struct {
	from  *Palette
	index int
}

// pair is a lamp and a material, by their places in one palette.
type pair struct {
	lamp, material int
}

// worked is one pair's product and its swatch, made once.
type worked struct {
	once  sync.Once
	light bands.Light
	seen  swatch.Swatch
}

// New is an empty palette, for one scene.
func New() *Palette {
	return new(Palette)
}

// Lamp gives the palette a light, and is how it is asked for after.
func (scene *Palette) Lamp(light bands.Light) Lamp {
	scene.lock.Lock()
	defer scene.lock.Unlock()
	scene.lamps = append(scene.lamps, light)
	return Lamp{from: scene, index: len(scene.lamps) - 1}
}

// Material gives the palette a share, and is how it is asked for after.
func (scene *Palette) Material(share bands.Share) Material {
	scene.lock.Lock()
	defer scene.lock.Unlock()
	scene.materials = append(scene.materials, share)
	return Material{from: scene, index: len(scene.materials) - 1}
}

// Light is the lamp's light through the material: bands.Lit, worked out
// the first time it is asked for and kept.
func (scene *Palette) Light(lamp Lamp, material Material) (bands.Light,
	error) {
	kept, err := scene.pair(lamp, material)
	if err != nil {
		var none bands.Light
		return none, err
	}
	return kept.light, nil
}

// Swatch is what the lamp's light through the material looks like: the
// swatch of Light, kept with it.
func (scene *Palette) Swatch(lamp Lamp, material Material) (swatch.Swatch,
	error) {
	kept, err := scene.pair(lamp, material)
	if err != nil {
		return swatch.NoLight(), err
	}
	return kept.seen, nil
}

// Worked is how many pairs the palette has worked out, which is how a
// render can tell whether keeping them is worth it.
//
//ninja:vow practice a render reports it; until one does, only tests ask
func (scene *Palette) Worked() int {
	return int(scene.count.Load())
}

// pair is the kept work for a lamp and a material, made the first time
// and read without a lock after. A lamp or a material this palette did
// not give out, or no palette at all, is refused.
func (scene *Palette) pair(lamp Lamp, material Material) (*worked, error) {
	if scene == nil || lamp.from != scene || material.from != scene {
		return nil, errors.New("palette: a lamp or a material " +
			"from another palette, or from none; ask the " +
			"palette that gave it out")
	}
	key := pair{lamp: lamp.index, material: material.index}
	found, held := scene.pairs.Load(key)
	if !held {
		found, _ = scene.pairs.LoadOrStore(key, new(worked))
	}
	kept := found.(*worked)
	kept.once.Do(func() {
		scene.lock.Lock()
		light := scene.lamps[lamp.index]
		share := scene.materials[material.index]
		scene.lock.Unlock()
		kept.light = bands.Lit(light, share)
		kept.seen = kept.light.Swatch()
		scene.count.Add(1)
	})
	return kept, nil
}
