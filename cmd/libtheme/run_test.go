package main

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/internal/host"
	"github.com/janearc/libtheme-css/internal/mocks"
)

// process is the program main runs in, for a test. It has a mock to stop at
// and the files its output went to.
type process struct {
	stopper   *mocks.MockStopper
	out, errs *os.File
}

// was is what a test changed about the program, and puts it back.
type was struct {
	arguments []string
	out, errs *os.File
	stopsWith host.Stopper
	made      process
}

// restore puts the program back as it was.
func (before was) restore() {
	os.Args, os.Stdout, os.Stderr, program = before.arguments, before.out,
		before.errs, before.stopsWith
	before.made.out.Close()
	before.made.errs.Close()
}

// emptyFile is a new empty file for a test's output.
func emptyFile(t *testing.T) *os.File {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "output-*")
	if err != nil {
		t.Fatal(err)
	}
	return file
}

// asProcess makes the program a test can call main in. It sets the arguments,
// sends the output to files, and holds the exit with a mock.
func asProcess(t *testing.T, arguments ...string) process {
	t.Helper()
	made := process{
		stopper: mocks.NewMockStopper(gomock.NewController(t)),
		out:     emptyFile(t),
		errs:    emptyFile(t),
	}
	before := was{os.Args, os.Stdout, os.Stderr, program, made}
	t.Cleanup(before.restore)
	os.Args = append([]string{"libtheme"}, arguments...)
	os.Stdout, os.Stderr, program = made.out, made.errs, made.stopper
	return made
}

// printed is what main wrote to a file.
func printed(t *testing.T, file *os.File) string {
	t.Helper()
	raw, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// call is a call of the command and what it did: the status it stopped with,
// and whether it said anything on each stream.
type call struct {
	arguments        []string
	status           int
	answers, laments bool
}

// The calls the command is held to.
var calls = []call{
	{[]string{"version"}, succeeded, true, false},
	{[]string{"--age"}, succeeded, true, false},
	{[]string{"known"}, succeeded, true, false},
	{[]string{"known", "--css"}, succeeded, true, false},
	{[]string{"roundtrip"}, succeeded, true, false},
	{[]string{"show", "#336699"}, succeeded, true, false},
	{[]string{"read", "testdata/dusk.css"}, succeeded, true, false},
	{[]string{"ramp", "#000000", "#ffffff"}, succeeded, true, false},
	{[]string{"ramp", "#000000", "#ffffff", "3"}, succeeded, true, false},
	{nil, wrongCall, false, true},
	{[]string{"nosuch"}, wrongCall, false, true},
	{[]string{"show"}, wrongCall, false, true},
	{[]string{"read"}, wrongCall, false, true},
	{[]string{"ramp", "#000000"}, wrongCall, false, true},
	{[]string{"show", "nonsense"}, failed, false, true},
	{[]string{"read", "testdata/nowhere.css"}, failed, false, true},
	{[]string{"ramp", "#000000", "#ffffff", "many"}, failed, false, true},
}

// main stops with 0 for a call that worked. It stops with 2 for a call it
// cannot make sense of, and shows the usage. It stops with 1 for a call that
// failed, and says why. The anti-tests are the calls that do not work. Each
// says its piece on the error stream and nothing on the output.
func TestMainStopsWithTheStatusOfTheCall(t *testing.T) {
	for _, each := range calls {
		t.Run(strings.Join(each.arguments, " "), each.holds)
	}
}

// holds runs main on the call and holds it to what the call should do.
func (each call) holds(t *testing.T) {
	made := asProcess(t, each.arguments...)
	made.stopper.EXPECT().Exit(each.status)
	main()
	if said := printed(t, made.out) != ""; said != each.answers {
		t.Errorf("answered %v, wanted %v", said, each.answers)
	}
	if said := printed(t, made.errs) != ""; said != each.laments {
		t.Errorf("lamented %v, wanted %v", said, each.laments)
	}
}

// aVerbless is a word made of letters. It is defined by this test.
const aVerbless = `[a-z]{1,8}`

// verbs are the words the command takes as its first. They are defined by
// dispatch.
var verbs = map[string]bool{"version": true, "known": true, "roundtrip": true,
	"show": true, "read": true, "ramp": true}

// The property: a first word that is not a verb is a call the command cannot
// make sense of, and the usage is shown.
func TestAnyOtherWordIsAWrongCall(t *testing.T) {
	made := asProcess(t)
	rapid.Check(t, wordIsWrong)
	if !strings.Contains(printed(t, made.errs), "libtheme known") {
		t.Error("the usage was not shown")
	}
}

// notAVerb is whether a word is one dispatch does not take.
func notAVerb(word string) bool { return !verbs[word] }

// wordIsWrong is one draw of the property above.
func wordIsWrong(draw *rapid.T) {
	word := rapid.StringMatching(aVerbless).Filter(notAVerb).Draw(draw, "word")
	if got := run([]string{word}); got != wrongCall {
		draw.Fatalf("%q stopped with %d", word, got)
	}
}
