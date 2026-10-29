package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// These tests hold what the mutation run found nothing holding. They cover
// the words of a changed line and the bound of a note. They cover a diff
// with one kind of change. They cover the files that are refused.

const (
	// shellDone is the status of a call that worked. It is defined by the
	// shell.
	shellDone = 0
	// shellFailed is the status of a call that failed. It is defined by the
	// shell.
	shellFailed = 1
	// shellWrong is the status of a call that makes no sense. It is defined
	// by the shell.
	shellWrong = 2
)

const (
	// wholeTurn is the degrees round the hue wheel. It is defined by the
	// circle.
	wholeTurn = 360
	// halfTurn is half of them. It is defined by the circle.
	halfTurn = 180
	// quarterTurn is a quarter of them. It is defined by the circle.
	quarterTurn = 90
	// aStep is a few degrees of hue. It is defined by this test.
	aStep = 10
	// half is a half. It is defined by halving.
	half = 0.5
)

// restorer puts a value back where it was.
type restorer[Value any] struct {
	place *Value
	was   Value
}

// put is the restoring.
func (before restorer[Value]) put() { *before.place = before.was }

// replacing puts a value in a place for the length of a test.
func replacing[Value any](t *testing.T, place *Value, value Value) {
	t.Helper()
	t.Cleanup(restorer[Value]{place, *place}.put)
	*place = value
}

// swatchHex is a swatch from a hex code.
func swatchHex(code string) swatch.Swatch { return srgb.MustHex(code).Swatch() }

// roleChange is a role that is in both sheets with the colours given.
func roleChange(name, before, after string) change {
	return change{name, swatchHex(before), swatchHex(after), true, true}
}

// The changed line says how far the role moved, times a hundred. Then it
// gives the change in lightness, chroma and hue. The numbers come from
// libtheme show. The two colours are oklch 50% 0.099 250.4 and 79% 0.143
// 131.0. Their oklab is 0.4993 -0.0330 -0.0930 and 0.7869 -0.0937 0.1080.
// The distance is 0.356.
func TestChangedLineSaysHowFarAndWhichWay(t *testing.T) {
	got := changedLine(roleChange("sky", "#336699", "#99cc66"), false)
	want := "~ sky                #336699 -> #99cc66  moved 35.6  " +
		"L +0.29 C +0.044 H -119"
	if got != want {
		t.Errorf("changed line\n got %q\nwant %q", got, want)
	}
}

// Black to white moves the whole span of oklab, which is 100.
func TestChangedLineMovedAHundredForTheWholeSpan(t *testing.T) {
	got := changedLine(roleChange("all", "#000000", "#ffffff"), false)
	if !strings.Contains(got, "moved 100.0 ") {
		t.Errorf("black to white reads %q", got)
	}
}

// turn is a hue before and after and the move between them.
type turn struct{ before, after, moved float64 }

// turns cross the top of the wheel both ways. They go up and down a quarter.
// They go not at all, and half way round.
var turns = []turn{
	{wholeTurn - aStep, aStep, aStep + aStep},
	{aStep, wholeTurn - aStep, -(aStep + aStep)},
	{nothing, quarterTurn, quarterTurn},
	{quarterTurn, nothing, -quarterTurn},
	{aStep, aStep, nothing},
	{nothing, halfTurn, -halfTurn},
}

// The hue moves the short way round. Half a turn reads as minus half.
// The pairs that do not wrap are the anti-test.
func TestHueMovesTheShortWay(t *testing.T) {
	for _, each := range turns {
		if got := hueMoved(each.before, each.after); got != each.moved {
			t.Errorf("%g to %g moved %g, wanted %g", each.before, each.after,
				got, each.moved)
		}
	}
}

// Only a move under what an eye can tell is noted. Exactly that is not.
// The anti-test is a move past it.
func TestUnseenIsUnderTheEye(t *testing.T) {
	if !unseen(ok.Eye * half) {
		t.Error("half of what an eye can tell was seen")
	}
	if unseen(ok.Eye) || unseen(ok.Eye+ok.Eye) {
		t.Error("what an eye can tell was unseen")
	}
}

// One kind of change alone still differs. That is a change, an addition or
// a removal. The anti-test is the same sheet twice.
func TestReportDiffersForEachKindAlone(t *testing.T) {
	base := sheetOf("a", "#101010")
	changed := sheetOf("a", "#f01010")
	added := sheetOf("a", "#101010", "b", "#202020")
	removed := sheetOf()
	for name, after := range map[string]*css.Sheet{
		"changed": changed, "added": added, "removed": removed} {
		var out bytes.Buffer
		if !report(&out, compared(base, after), false) {
			t.Errorf("only %s: the diff said nothing differed", name)
		}
	}
	var out bytes.Buffer
	if report(&out, compared(base, base), false) {
		t.Error("the same sheet differed")
	}
}

// aFile is a file to stand for stdout, and whether terminal should say it
// is a terminal.
type aFile struct {
	file *os.File
	want bool
}

// opened is a file opened for writing, closed when the test ends.
func opened(t *testing.T, path string) *os.File {
	t.Helper()
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, private)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closer{file}.close)
	return file
}

// closer closes a file.
type closer struct{ file *os.File }

// close closes the file, and does not mind that it is closed already.
func (each closer) close() { each.file.Close() }

// terminal is true for a character device like a screen. It is false for a
// file and for a stdout that cannot be asked. The anti-tests are those two.
func TestTerminalIsACharacterDeviceAndNothingElse(t *testing.T) {
	where := t.TempDir()
	closed := opened(t, filepath.Join(where, "closed"))
	closed.Close()
	for name, each := range map[string]aFile{
		"device": {opened(t, os.DevNull), true},
		"file":   {opened(t, filepath.Join(where, "out")), false},
		"closed": {closed, false},
	} {
		replacing(t, &os.Stdout, each.file)
		if got := terminal(); got != each.want {
			t.Errorf("%s: terminal said %v", name, got)
		}
	}
}

// statusCall is a call and the status it stops with.
type statusCall struct {
	status    int
	arguments []string
}

// The statuses are the shell's numbers. They are done, failed and a call
// that makes no sense.
func TestStatusesAreTheShellsNumbers(t *testing.T) {
	gone := filepath.Join(t.TempDir(), "gone")
	for _, each := range []statusCall{
		{shellDone, []string{"--age"}},
		{shellFailed, []string{"show", gone}},
		{shellWrong, []string{"nosuch"}},
	} {
		if got := called(each.arguments...).status; got != each.status {
			t.Errorf("%v stopped with %d, wanted %d", each.arguments, got,
				each.status)
		}
	}
}

// show writes a role's lightness as a percent: white is 100%.
func TestShowWritesLightnessAsAPercent(t *testing.T) {
	path := stored(t, "white.css", ":root { --ink: #ffffff; }\n")
	var out bytes.Buffer
	if err := show([]string{path}, streams{out: &out}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "#ffffff  oklch(100% 0.000 ") {
		t.Errorf("show wrote %q", out.String())
	}
}

// Reading a bad terminal theme gives an error and no roles.
func TestReadTerminalRefusesABadTheme(t *testing.T) {
	roles, err := readTerminal(side{}, []byte("palette = 300=#000000\n"))
	if err == nil || roles != nil {
		t.Errorf("a bad theme gave %v, %v", roles, err)
	}
}

// git's /dev/null is no roles for a file of any dialect, though a reader of
// that dialect would refuse an empty file.
func TestDevNullIsNoRolesInAnyDialect(t *testing.T) {
	for _, name := range []string{"a.css", "a.lua", "a.json", "a"} {
		roles, err := rolesOf(side{name: name, path: os.DevNull})
		if err != nil || len(roles.Names()) != nothing {
			t.Errorf("%s from /dev/null gave %v, %v", name, roles, err)
		}
	}
}

// Missing policy files are errors that say so. The reference range needs
// no file and is the anti-test.
func TestPolicyFromAMissingFileFails(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nowhere.json")
	if _, err := policyFrom(missing); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a missing policy gave %v", err)
	}
	if _, err := policyFrom(""); err != nil {
		t.Errorf("the reference range gave %v", err)
	}
}

// Sheet fails on a missing file for each program. It says the file is not
// there.
func TestSheetFailsOnAMissingFile(t *testing.T) {
	gone := filepath.Join(t.TempDir(), "gone")
	for _, option := range sheetOptions {
		to := streams{out: &bytes.Buffer{}, errs: &bytes.Buffer{}}
		if err := sheet([]string{option, gone}, to); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s with no file gave %v", option, err)
		}
	}
}

// The opening comment of a theme ends at the first line that is not one.
// The anti-test is a comment later in the file, which is not taken.
func TestCommentsStopAtTheFirstLineThatIsNotOne(t *testing.T) {
	text := "# one\n# two\nforeground = #fff\n# three\nbackground = #000\n"
	if got := comments(text); got != "one\ntwo" {
		t.Errorf("comments gave %q", got)
	}
}

// blackAt makes claude's black slot sit at a palette number for the length
// of a test.
func blackAt(t *testing.T, number int) {
	t.Helper()
	slots := slices.Clone(claude.Slots)
	for position := range slots {
		if slots[position].Role == "black" {
			slots[position].Number = number
		}
	}
	replacing(t, &claude.Slots, slots)
}

// The first palette number that is claude's rather than the terminal's own.
// It is defined by the claude dialect: slots below it are skipped.
const firstOfClaudes = 16

// lastOfTheTerminals is the palette number just under it. It is defined by
// the same.
const lastOfTheTerminals = 15

// Palette numbers from 16 up are claude's. Those below are the terminal's.
// The slot at 16 is claude's and the one at 15 is not.
func TestClaudeNumberStartsAtSixteen(t *testing.T) {
	blackAt(t, firstOfClaudes)
	if !claudeNumber("palette-16") {
		t.Error("palette-16 is not claude's")
	}
	blackAt(t, lastOfTheTerminals)
	if claudeNumber("palette-15") || claudeNumber("ink") {
		t.Error("a palette number below 16 or a name was taken as claude's")
	}
}
