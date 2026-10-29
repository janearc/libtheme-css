package palette

import (
	"fmt"
	"math"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// Steps is how many points a table holds, evenly from 0 to 1, ends
// included. The caller chooses: more is closer, and costs more to build.
type Steps int

// JND is a size of error in just-noticeable differences: ok.Eye, 0.02 in
// oklab, is one. Under one, most people see no difference side by side.
type JND float64

// checked is how many points a table's error is measured at, between each
// pair of its own.
const checked = 16

// least is the fewest steps a table can have: one at each end.
const least = 2

// enough refuses a table too small to hold both ends.
func enough(steps Steps) error {
	if steps < least {
		return fmt.Errorf("palette: a table needs at least %d "+
			"steps, one at each end, and %d is fewer", least, steps)
	}
	return nil
}

// Curve is a table for a curve from 0 to 1: read between its points by a
// straight line, exact at the points, and exact outside 0 to 1, where it
// asks the curve itself.
type Curve struct {
	exact  func(float64) float64
	points []float64
	worst  JND
}

// ToLinear is a table for srgb's curve from a lamp's level to linear
// light, srgb.ToLinear.
func ToLinear(steps Steps) (*Curve, error) {
	looks := func(light float64) swatch.Swatch {
		return srgb.FromLight(light, light, light)
	}
	return newCurve(steps, srgb.ToLinear, looks)
}

// FromLinear is a table for srgb's curve from linear light to a lamp's
// level, srgb.FromLinear.
func FromLinear(steps Steps) (*Curve, error) {
	looks := func(level float64) swatch.Swatch {
		light := srgb.ToLinear(level)
		return srgb.FromLight(light, light, light)
	}
	return newCurve(steps, srgb.FromLinear, looks)
}

// newCurve is a table of exact at steps points, and its error measured:
// what its answer looks like, as a grey, beside what the curve's does.
func newCurve(steps Steps, exact func(float64) float64,
	looks func(float64) swatch.Swatch) (*Curve, error) {
	if err := enough(steps); err != nil {
		return nil, err
	}
	table := &Curve{exact: exact, points: make([]float64, steps)}
	for index := range table.points {
		table.points[index] = exact(float64(index) / float64(steps-1))
	}
	samples := (int(steps)-1)*checked + 1
	for index := range samples {
		at := float64(index) / float64(samples-1)
		table.worst = max(table.worst, apart(looks(table.At(at)),
			looks(exact(at))))
	}
	return table, nil
}

// At is the curve at a value: between two points, the line between them.
func (table *Curve) At(value float64) float64 {
	if !(value >= 0 && value <= 1) { // NaN too: it goes to the curve
		return table.exact(value)
	}
	index, along := place(value, len(table.points))
	if along == 0 {
		return table.points[index]
	}
	low, high := table.points[index], table.points[index+1]
	return low + (high-low)*along
}

// Worst is the table's worst error found when it was made: the most at
// checked points between each pair of its own, which is a sampled bound,
// not a proof.
func (table *Curve) Worst() JND { return table.worst }

// place is where a value from 0 to 1 falls among count points: the point
// at or below it, and how far it is on toward the next. At 1 that is the
// last point and nothing further, so a caller never reads past it.
func place(value float64, count int) (int, float64) {
	position := value * float64(count-1)
	index := int(math.Floor(position))
	return index, position - float64(index)
}

// apart is how different two swatches look, in JNDs.
func apart(table, curve swatch.Swatch) JND {
	return JND(ok.Distance(ok.FromSwatch(table), ok.FromSwatch(curve)) /
		ok.Eye)
}

// Ramp is a table for a blend between two fixed colours along t from 0
// to 1, by a mixer: read between its points by a straight line in XYZ,
// exact at the points, and exact outside 0 to 1, where it asks the mixer.
type Ramp struct {
	mixer    functions.Mixer
	from, to swatch.Swatch
	points   []swatch.Swatch
	worst    JND
}

// Blend is a table for the mixer's blend from one colour to another.
func Blend(from, to swatch.Swatch, mixer functions.Mixer,
	steps Steps) (*Ramp, error) {
	if err := enough(steps); err != nil {
		return nil, err
	}
	ramp := &Ramp{mixer: mixer, from: from, to: to,
		points: make([]swatch.Swatch, steps)}
	for index := range ramp.points {
		ramp.points[index] = mixer.Mix(from, to,
			float64(index)/float64(steps-1))
	}
	samples := (int(steps)-1)*checked + 1
	for index := range samples {
		along := float64(index) / float64(samples-1)
		ramp.worst = max(ramp.worst, apart(ramp.At(along),
			mixer.Mix(from, to, along)))
	}
	return ramp, nil
}

// At is the blend at t: between two points, the line between them.
func (ramp *Ramp) At(along float64) swatch.Swatch {
	if !(along >= 0 && along <= 1) { // NaN too: it goes to the mixer
		return ramp.mixer.Mix(ramp.from, ramp.to, along)
	}
	index, between := place(along, len(ramp.points))
	if between == 0 {
		return ramp.points[index]
	}
	lx, ly, lz := ramp.points[index].XYZ()
	hx, hy, hz := ramp.points[index+1].XYZ()
	return swatch.FromXYZ(lx+(hx-lx)*between, ly+(hy-ly)*between,
		lz+(hz-lz)*between)
}

// Worst is the table's worst error found when it was made: the most at
// checked points between each pair of its own, which is a sampled bound,
// not a proof.
func (ramp *Ramp) Worst() JND { return ramp.worst }
