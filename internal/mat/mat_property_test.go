package mat_test

import (
	"math"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/internal/mat"
)

// The numbers the tests pick.
const (
	// offLeast is the least an entry off the diagonal of a drawn matrix is.
	// It is defined by this test.
	offLeast = -1.0
	// offMost is the most an entry off the diagonal is. It is defined by
	// this test.
	offMost = 1.0
	// onLeast is the least an entry on the diagonal is. It is defined by
	// this test: with the off-diagonals it makes the matrix dominated by
	// its diagonal, so it is always invertible and well conditioned.
	onLeast = 4.0
	// onMost is the most an entry on the diagonal is. It is defined by this
	// test.
	onMost = 8.0
	// valueLeast is the least a drawn column value or scale is. It is
	// defined by this test.
	valueLeast = -3.0
	// valueMost is the most a drawn column value or scale is. It is defined
	// by this test.
	valueMost = 3.0
	// tolerance is how far two results may differ. It is defined by this
	// test: the matrices are well conditioned, so the error is far below it.
	tolerance = 1e-9
	// order is the rows and columns of a matrix. It is defined by M.
	order = 3
	// keep is a scale that leaves a column as it is. It is defined by
	// multiplication.
	keep = 1.0
	// anySign asks math.IsInf for an infinity of either sign. It is
	// defined by math.IsInf.
	anySign = 0
	// lose is a scale that takes a column away. It is defined by
	// multiplication.
	lose = 0.0
	// second is the position of the second row. It is defined by M.
	second = 1
	// third is the position of the third row. It is defined by M.
	third = 2
)

// dominant is a drawn matrix whose diagonal outweighs the rest.
func dominant(t *rapid.T) mat.M {
	off := rapid.Float64Range(offLeast, offMost)
	on := rapid.Float64Range(onLeast, onMost)
	var matrix mat.M
	for row := range order {
		for column := range order {
			if row == column {
				matrix[row][column] = on.Draw(t, "diagonal")
			} else {
				matrix[row][column] = off.Draw(t, "off")
			}
		}
	}
	return matrix
}

// near is whether two numbers agree within the tolerance.
func near(left, right float64) bool {
	return math.Abs(left-right) <= tolerance
}

// The property: the inverse of a matrix undoes it.
func TestInverseUndoesTheMatrix(t *testing.T) {
	rapid.Check(t, inverseUndoes)
}

// inverseUndoes is one draw of the property above.
func inverseUndoes(t *rapid.T) {
	matrix := dominant(t)
	value := rapid.Float64Range(valueLeast, valueMost)
	x, y, z := value.Draw(t, "x"), value.Draw(t, "y"), value.Draw(t, "z")
	one, two, three := matrix.Apply(x, y, z)
	back1, back2, back3 := matrix.Inverse().Apply(one, two, three)
	if !near(back1, x) || !near(back2, y) || !near(back3, z) {
		t.Fatalf("%v then its inverse gave %v %v %v, not %v %v %v", matrix,
			back1, back2, back3, x, y, z)
	}
}

// The property: scaling the columns of a matrix is applying it to the
// inputs scaled by the same amounts.
func TestScaleIsTheMatrixOfScaledInputs(t *testing.T) {
	rapid.Check(t, scalesTheInputs)
}

// scalesTheInputs is one draw of the property above.
func scalesTheInputs(t *rapid.T) {
	matrix := dominant(t)
	value := rapid.Float64Range(valueLeast, valueMost)
	s1, s2, s3 := value.Draw(t, "s1"), value.Draw(t, "s2"), value.Draw(t, "s3")
	x, y, z := value.Draw(t, "x"), value.Draw(t, "y"), value.Draw(t, "z")
	got1, got2, got3 := matrix.Scale(s1, s2, s3).Apply(x, y, z)
	want1, want2, want3 := matrix.Apply(s1*x, s2*y, s3*z)
	if !near(got1, want1) || !near(got2, want2) || !near(got3, want3) {
		t.Fatalf("scaled %v by %v %v %v: %v %v %v, want %v %v %v", matrix,
			s1, s2, s3, got1, got2, got3, want1, want2, want3)
	}
}

// hasInverse is whether every entry of the inverse is a finite number.
func hasInverse(matrix mat.M) bool {
	for _, row := range matrix.Inverse() {
		for _, entry := range row {
			if math.IsInf(entry, anySign) || math.IsNaN(entry) {
				return false
			}
		}
	}
	return true
}

// The anti-test: a matrix with a row of nothing is one nothing undoes, and
// Inverse does not pretend to an inverse. Its entries are not finite, so a
// caller can see the matrix could not be inverted.
func TestASingularMatrixHasNoFiniteInverse(t *testing.T) {
	rapid.Check(t, singularHasNone)
}

// singularHasNone is one draw of the property above.
func singularHasNone(t *rapid.T) {
	matrix := dominant(t)
	if !hasInverse(matrix) {
		t.Fatalf("a good matrix had no inverse: %v", matrix)
	}
	matrix[second] = [order]float64{}
	if hasInverse(matrix) {
		t.Fatalf("a matrix with a row of nothing had an inverse: %v", matrix)
	}
}

// The anti-test: scaling a column by nothing takes the column away, and the
// matrix that is left has no inverse either.
func TestScalingAColumnByNothingLosesTheInverse(t *testing.T) {
	rapid.Check(t, losesTheInverse)
}

// losesTheInverse is one draw of the property above.
func losesTheInverse(t *rapid.T) {
	matrix := dominant(t)
	flat := matrix.Scale(keep, lose, keep)
	if hasInverse(flat) {
		t.Fatalf("%v with a column taken away had an inverse", matrix)
	}
	same := matrix.Scale(keep, keep, keep)
	if same != matrix || matrix[third][third] != same[third][third] {
		t.Fatalf("scaling by one changed the matrix: %v", same)
	}
}
