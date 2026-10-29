package glamour_test

import (
	"testing"

	"github.com/janearc/libtheme-css/dialects/glamour"
)

// Read refuses a colour key that holds another kind of value. The
// message names the key and the kind, and says it is not a colour.
func TestAWrongKindIsNamedAsNotAColour(t *testing.T) {
	for kind, value := range map[string]string{"a number": "42",
		"null": "null", "true or false": "true", "a list": `["#ffffff"]`,
		"an object": `{"a": 1}`} {
		_, err := glamour.Read([]byte(`{"document": {"color": ` + value + `}}`))
		want := "glamour: document.color is " + kind + ", not a colour"
		if err == nil || err.Error() != want {
			t.Errorf("%s gave %v, want %q", kind, err, want)
		}
	}
}
