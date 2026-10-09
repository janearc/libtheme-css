package age_test

import (
	"testing"
	"time"

	"github.com/janearc/libtheme-css/internal/age"
)

func TestOf(t *testing.T) {
	now := time.Date(2026, 10, 8, 23, 0, 0, 0, time.UTC)
	cases := []struct {
		build, built, want string
	}{
		{"dev", "", "spectra, built by go rather than by game: " +
			"no commit and no commit time in it"},
		{"", "", "spectra, built by go rather than by game: " +
			"no commit and no commit time in it"},
		{"284aab1", "", "spectra 284aab1, with no commit time in it"},
		{"284aab1", "yesterday", "spectra 284aab1, committed at yesterday"},
		{"284aab1-dirty", "2026-10-08T21:30:00Z",
			"spectra 284aab1-dirty, committed 1h30m0s ago"},
	}
	for _, c := range cases {
		got := age.Of("spectra", c.build, c.built, now)
		if got != c.want {
			t.Errorf("Of(%q, %q) = %q, want %q",
				c.build, c.built, got, c.want)
		}
	}
}
