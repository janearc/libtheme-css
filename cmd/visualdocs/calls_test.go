package main

import (
	"embed"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// call is a call of the command and what it did. It says the status it
// stopped with, a piece of what it printed, a piece of what it complained of,
// and a piece that it did not print.
type call struct {
	arguments []string
	words     string
	status    int
	said      string
	complains string
	lacks     string
}

// Enter, typed eight times: enough to page through every page.
const enters = "\n\n\n\n\n\n\n\n"

// The calls the command is held to. Thirteen work, one of them stopped by q
// after the first page. The last three are the anti-tests: a call it cannot
// make sense of says so on the error stream and prints nothing.
var calls = []call{
	{[]string{"version"}, "", succeeded, "visualdocs", "", ""},
	{[]string{"--age"}, "", succeeded, "visualdocs", "", ""},
	{[]string{"swatch"}, "", succeeded, "== swatch", "", "== observer"},
	{[]string{"swatch", "--css"}, "", succeeded, "/* swatch */", "", "=="},
	{[]string{"--css", "swatch"}, "", succeeded, "/* swatch */", "", "=="},
	{[]string{"swatch", "--go"}, "", succeeded, "// page_swatch.go", "", "=="},
	{[]string{"--css"}, "", succeeded, "/* css */", "", "[1/8]"},
	{[]string{"--go"}, "", succeeded, "// page_css.go", "", "[1/8]"},
	{nil, enters, succeeded, "[1/8] enter for observer, q to stop: ", "", ""},
	{nil, enters, succeeded, "== css", "", ""},
	{nil, "", succeeded, "== css", "", ""},
	{nil, "q\n", succeeded, "[1/8]", "", "== observer"},
	{[]string{"swatch", "ok"}, "", succeeded, "== ok", "", "== swatch"},
	{[]string{"--bogus"}, "", wrongCall, "",
		`modes are --css and --go, not "--bogus"`, ""},
	{[]string{"swatch", "--bogus"}, "", wrongCall, "",
		`modes are --css and --go, not "--bogus"`, ""},
	{[]string{"nosuch"}, "", wrongCall, "",
		`no page called "nosuch"; there is swatch, observer, bands, ok,`, ""},
}

// main stops with the status of the call, and holds each call to what it
// printed and what it complained of.
func TestMainStopsWithTheStatusOfTheCall(t *testing.T) {
	for _, each := range calls {
		name := strings.Join(each.arguments, " ") + "|" + each.words
		t.Run(name, each.holds)
	}
}

// holds runs main on the call and holds it to what the call should do.
func (each call) holds(t *testing.T) {
	made := asProcess(t, each.words, each.arguments...)
	made.stopper.EXPECT().Exit(each.status)
	main()
	out, errs := printed(t, made.out), printed(t, made.errs)
	if !strings.Contains(out, each.said) || !strings.Contains(errs, each.complains) {
		t.Errorf("printed %q, complained %q", out, errs)
	}
	if each.lacks != "" && strings.Contains(out, each.lacks) {
		t.Errorf("printed %q, which holds %q", out, each.lacks)
	}
	if each.complains != "" && out != "" || each.complains == "" && errs != "" {
		t.Errorf("printed %q and complained %q", out, errs)
	}
}

// sourcesWas is the embedded source a test replaced.
type sourcesWas struct{ files embed.FS }

// put gives the source back.
func (was sourcesWas) put() { sources = was.files }

// A page whose source is not embedded cannot be shown as go. The call fails
// with the error and prints none of the page. The same call with the source
// there is in the calls above, and prints it.
func TestGoFailsWhenTheSourceIsNotThere(t *testing.T) {
	t.Cleanup(sourcesWas{sources}.put)
	sources = embed.FS{}
	made := asProcess(t, "", "swatch", "--go")
	made.stopper.EXPECT().Exit(failed)
	main()
	if out := printed(t, made.out); out != "" {
		t.Errorf("printed %q", out)
	}
	if errs := printed(t, made.errs); !strings.Contains(errs, "visualdocs:") {
		t.Errorf("complained %q", errs)
	}
}

// aWord is a word made of letters. It is defined by this test.
const aWord = `[a-z]{1,8}`

// pageNamed is whether a word names a page.
func pageNamed(word string) bool {
	for _, each := range pages {
		if each.name == word {
			return true
		}
	}
	return false
}

// notAPage is whether a word does not name a page.
func notAPage(word string) bool { return !pageNamed(word) }

// notAMode is whether a word, written as an option, is neither of the modes
// nor the version call.
func notAMode(word string) bool {
	return word != "css" && word != "go" && word != "age"
}

// The property: a word that names no page is a wrong call, and so is an
// option that is no mode.
func TestWhatIsNotAPageOrAModeIsAWrongCall(t *testing.T) {
	made := asProcess(t, "")
	rapid.Check(t, wrongWord)
	if printed(t, made.out) != "" {
		t.Error("a wrong call printed")
	}
}

// wrongWord is one draw of the property above.
func wrongWord(draw *rapid.T) {
	word := rapid.StringMatching(aWord).Filter(notAPage).Draw(draw, "word")
	option := rapid.StringMatching(aWord).Filter(notAMode).Draw(draw, "option")
	if got := run([]string{word}); got != wrongCall {
		draw.Fatalf("page %q stopped with %d", word, got)
	}
	if got := run([]string{"--" + option}); got != wrongCall {
		draw.Fatalf("mode --%s stopped with %d", option, got)
	}
}
