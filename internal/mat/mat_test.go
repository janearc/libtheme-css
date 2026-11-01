package mat

import (
	"math"
	"testing"
)

// sameEverywhere is a matrix and its inverse as amd64 computes it, which
// never fuses a multiply into an add. arm64 may, so the same bits on a
// mac say each product is rounded on its own there too. Each matrix was
// picked because arm64 gives other bits when one rounding is taken out:
// OKLab's in the minors, Bradford's in the determinant's first and
// second products, near-singular in its third.
var sameEverywhere = []struct {
	name    string
	matrix  M
	inverse [3][3]uint64
}{
	{"OKLab XYZ to LMS", M{
		{0.8189330101, 0.3618667424, -0.1288597137},
		{0.0329845436, 0.9293118715, 0.0361456387},
		{0.0482003018, 0.2643662691, 0.6338517070},
	}, [3][3]uint64{
		{0x3ff3a1d946a3a87f, 0xbfe1d97f58537d01, 0x3fd20019ca670855},
		{0xbfa4c6ecd6633e57, 0x3ff1cbcddbfc1706, 0xbfb259671ec13145},
		{0xbfb38db94efa7dd7, 0xbfdaf98f8c4a28d6, 0x3ff960ecaf5e947e},
	}},
	{"Bradford", M{
		{0.8951, 0.2664, -0.1614},
		{-0.7502, 1.7135, 0.0367},
		{0.0389, -0.0685, 1.0296},
	}, [3][3]uint64{
		{0x3fef9572254ba231, 0xbfc2d2ac83086a84, 0x3fc479a7fabd13c3},
		{0x3fdbaae3b8d66d96, 0x3fe0966847b979f9, 0x3fa93cb32a4eca52},
		{0xbf817779fb6e7848, 0x3fa4807e22e340fb, 0x3feefdd7cfa0989b},
	}},
	{"near-singular", M{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
		{0.7, 0.8, 0.91},
	}, [3][3]uint64{
		{0x4054d55555555680, 0xc0682aaaaaaaac0b, 0x4059000000000170},
		{0xc0675555555556a6, 0x4078caaaaaaaac17, 0xc069000000000170},
		{0x405900000000015c, 0xc06900000000016d, 0x4059000000000173},
	}},
}

// The inverse is the same to the bit on every architecture.
func TestInverseIsTheSameOnEveryMachine(t *testing.T) {
	for _, each := range sameEverywhere {
		inv := each.matrix.Inverse()
		for i := range 3 {
			for j := range 3 {
				got := math.Float64bits(inv[i][j])
				if got != each.inverse[i][j] {
					t.Errorf("%s: inverse[%d][%d] = %#x, want %#x",
						each.name, i, j, got, each.inverse[i][j])
				}
			}
		}
	}
}

// A matrix times its inverse is the identity, to the float.
func TestInverse(t *testing.T) {
	m := M{{2, 1, 0}, {1, 3, 1}, {0, 1, 4}}
	inv := m.Inverse()
	for i := 0; i < 3; i++ {
		a, b, c := inv.Apply(m[0][i], m[1][i], m[2][i])
		got := [3]float64{a, b, c}
		for j := 0; j < 3; j++ {
			want := 0.0
			if i == j {
				want = 1
			}
			if math.Abs(got[j]-want) > 1e-12 {
				t.Errorf("inverse * m column %d = %v", i, got)
			}
		}
	}
}
