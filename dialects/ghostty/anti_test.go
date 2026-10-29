package ghostty_test

import (
	"regexp"
	"testing"

	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// The numbers the anti-tests pick.
const (
	// beyondTheLamps is an amount of light below none, which no lamp can
	// make. It is defined by this test.
	beyondTheLamps = -0.5
	// brighterThanWhite is more light than the white has. It is defined by
	// this test.
	brighterThanWhite = 4.0
	// fullLight is an amount of light at the white. It is defined by the
	// scale: the white is one.
	fullLight = 1.0
)

// aGround is a theme with only its ground.
const aGround = "background = #808a63\n"

// The anti-test: lines that are not colour settings are passed over, and
// nothing is made of them. Font sizes, keybinds, lines with no equals sign
// and keys nobody has are no error and no colour.
func TestReadPassesOverWhatIsNotAColour(t *testing.T) {
	text := "font-size = 12\nkeybind = ctrl+a=new_tab\nno equals sign\n" +
		"invented-color = #123456\n" + aGround
	theme, err := read(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(theme.Settings) != one || len(theme.Palette) != nothing {
		t.Errorf("settings %v, palette %v", theme.Settings, theme.Palette)
	}
}

// light is amounts of light from the three lamps.
type light struct{ red, green, blue float64 }

// hexColour is what Hex writes: a hash and six hex digits.
var hexColour = regexp.MustCompile(`^#[0-9a-f]{6}$`)

// The anti-test: a colour no lamp can make is still written as a hex colour
// and not as something a theme file would reject.
func TestHexWritesAColourForLightNoLampCanMake(t *testing.T) {
	for _, amounts := range []light{
		{beyondTheLamps, fullLight, nothing},
		{fullLight, fullLight, brighterThanWhite},
	} {
		got := ghostty.Hex(srgb.FromLight(amounts.red, amounts.green,
			amounts.blue))
		if !hexColour.MatchString(got) {
			t.Errorf("light %v was written as %q", amounts, got)
		}
	}
}
