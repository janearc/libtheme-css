package glamour_test

import (
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/dialects/glamour"
)

// The anti-tests: a colour key that is there and is not a string is a
// style that is wrong, not one with the colour left out. Each of JSON's
// other kinds of value is refused, and the key is named.
func TestAColourThatIsNotAStringIsRefused(t *testing.T) {
	for name, value := range map[string]string{"number": "42", "null": "null",
		"object": `{"a": 1}`, "array": `["#ffffff"]`, "true": "true",
		"false": "false"} {
		_, err := glamour.Read([]byte(`{"document": {"color": ` + value + `}}`))
		if err == nil || !strings.Contains(err.Error(), "document.color") {
			t.Errorf("a %s colour gave %v", name, err)
		}
	}
}

// A key whose parent is there and is not an object is refused the same way:
// the parent is named.
func TestAParentThatIsNotAnObjectIsRefused(t *testing.T) {
	for name, value := range map[string]string{"number": "42", "null": "null",
		"string": `"#ffffff"`, "array": "[]"} {
		_, err := glamour.Read([]byte(`{"document": ` + value + `}`))
		if err == nil || !strings.Contains(err.Error(), "document") {
			t.Errorf("a %s parent gave %v", name, err)
		}
	}
	if _, err := glamour.Read([]byte(`{"code_block": {"chroma": 5}}`)); err == nil {
		t.Error("a chroma that is a number was read")
	}
}

// What is absent is fine: a style need not set every colour. Nothing, an
// empty parent and a style with one colour all read, the last as one role.
func TestWhatIsAbsentIsFine(t *testing.T) {
	for name, style := range map[string]string{"nothing": `{}`,
		"empty parent": `{"document": {}}`,
		"other keys":   `{"document": {"margin": 2}, "heading": {"bold": true}}`} {
		roles, err := glamour.Read([]byte(style))
		if err != nil || len(roles.Names()) != nothing {
			t.Errorf("%s read as %v, %v", name, roles.Names(), err)
		}
	}
	roles, err := glamour.Read([]byte(`{"heading": {"color": "#ffffff"}}`))
	if err != nil || len(roles.Names()) != one {
		t.Errorf("one colour read as %v, %v", roles.Names(), err)
	}
}

// nothing and one are counts of roles. They are defined by counting.
const (
	// nothing is no roles. It is defined by counting.
	nothing = 0
	// one is a single role. It is defined by counting.
	one = 1
)

// The property: a value at a colour key is read exactly when it is a
// string of hex, whatever other kind of value it is.
func TestOnlyAStringIsAColour(t *testing.T) {
	rapid.Check(t, onlyAString)
}

// onlyAString is one draw of the property above.
func onlyAString(t *rapid.T) {
	kinds := []string{`"#ffffff"`, `42`, `null`, `true`, `[]`, `{}`, `[1]`, `{"a": {}}`, `"nonsense"`}
	value := rapid.SampledFrom(kinds).Draw(t, "value")
	_, err := glamour.Read([]byte(`{"heading": {"color": ` + value + `}}`))
	if (err == nil) != (value == kinds[firstKind]) {
		t.Fatalf("value %s: error %v", value, err)
	}
}

// firstKind is where the good colour is in the kinds. It is defined by the
// list above it.
const firstKind = 0

// Text that is not json at all is refused, with the reader's name on it.
func TestNotJSONIsRefused(t *testing.T) {
	for _, text := range []string{"{", "not json", `["a list"]`} {
		if _, err := glamour.Read([]byte(text)); err == nil {
			t.Errorf("%q was read as a style", text)
		}
	}
}

// Glamour reads a bare number as one of the 256 terminal colours. Read as
// hex, "252" would be #225522. That is a dark green where glamour draws a
// light grey. Colours with no # are refused. The complaint names the key and
// the text. The anti-tests are digits with the # (read) and a hex code that
// is not one (refused as before).
func TestABareNumberIsNotHex(t *testing.T) {
	for _, text := range []string{"252", "39", "ffffff", "fff"} {
		_, err := glamour.Read([]byte(`{"document": {"color": "` + text + `"}}`))
		want := `document.color: "` + text + `" is not a #rrggbb`
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q gave %v, want %q", text, err, want)
		}
	}
	roles, err := glamour.Read([]byte(`{"document": {"color": "#252525"}}`))
	if err != nil || len(roles.Names()) != one {
		t.Errorf("#252525 read as %v, %v", roles.Names(), err)
	}
	_, err = glamour.Read([]byte(`{"document": {"color": "#nonsense"}}`))
	if err == nil || !strings.Contains(err.Error(), "document.color") {
		t.Errorf("#nonsense gave %v", err)
	}
}
