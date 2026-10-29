//ninja:vow dojo onASurface answers differently for prose on a ground that is no surface, which no group in the shipped groups file draws, so only a direct call reaches it.
package readable

import "testing"

// onASurface holds only when the text is prose or state and the ground
// is one of the surfaces.
func TestOnASurfaceNeedsTextAndASurface(t *testing.T) {
	held := gate{}
	for name, one := range map[string]struct {
		ink, ground string
		want        bool
	}{
		"prose on the ground":    {"ink", "bg", true},
		"state on the ground":    {"err", "bg", true},
		"state on a panel":       {"hint", "panel", true},
		"prose on the selection": {"code", "visual", true},
		"prose on a non-surface": {"ink", "string", false},
		"state on a non-surface": {"err", "heading", false},
		"ground on the ground":   {"bg", "bg", false},
		"furniture on a panel":   {"gutter", "panel", false},
	} {
		if got := held.onASurface(one.ink, one.ground); got != one.want {
			t.Errorf("%s: %v, want %v", name, got, one.want)
		}
	}
}
