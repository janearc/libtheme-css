package ghostty_test

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// grey is the one colour every role of the test theme has.
const grey = "#808080"

// A header's every kind of line end starts another comment line, so a
// header cannot set a ghostty key.
func TestAHeaderCannotLeaveItsComment(t *testing.T) {
	roles := css.New()
	roles.Set("ground", srgb.MustHex(grey).Swatch())
	roles.Set("ink", srgb.MustHex(grey).Swatch())
	for _, name := range ghostty.Ansi {
		roles.Set(name, srgb.MustHex(grey).Swatch())
	}
	theme, err := ghostty.Of(roles)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	header := "one\rpalette = 1=#000000\r\ntwo\vkeybind = x\u2028background = #000000"
	if err := theme.Write(&out, header); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.Contains(line, "keybind") && !strings.HasPrefix(line, "#") {
			t.Errorf("header text became a setting: %q", line)
		}
	}
	if strings.ContainsAny(out.String(), "\r\v\f\u2028") {
		t.Error("a line break other than a newline reached the file")
	}
}
