// Package colourway is tested from inside here.
//
//ninja:vow dojo these tests reach copy and writeFile, which the package does not export
package colourway

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/internal/mocks"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// These tests hold what the mutation run found nothing holding. They cover
// the modes of what is written. They cover the edge of an opening comment
// and what an install records as replaced. They cover the three grounds.

// umasked is a umask to put back.
type umasked struct{ was int }

// put sets the umask back.
func (before umasked) put() { syscall.Umask(before.was) }

// withoutAUmask clears the umask for the test. Files get the mode they ask
// for.
func withoutAUmask(t *testing.T) {
	t.Helper()
	t.Cleanup(umasked{syscall.Umask(nothing)}.put)
}

const (
	// aFileMode is the mode Write and writeFile give a file. It is defined
	// by them.
	aFileMode fs.FileMode = 0o644
	// aFolderMode is the mode writeFile gives a folder. It is defined by
	// writeFile.
	aFolderMode fs.FileMode = 0o755
	// twoFiles is how many files the replacing test writes. It is defined
	// by the test.
	twoFiles = 2
	// closeEnough is how near a share must come to its picked number. It
	// is defined by this test: a millionth.
	closeEnough = 1e-6
)

// writesWords writes a few words.
func writesWords(out io.Writer) error {
	_, err := out.Write([]byte("words"))
	return err
}

// writesNothing writes nothing.
func writesNothing(io.Writer) error { return nil }

// modeOf is the permission bits of a path.
func modeOf(t *testing.T, path string) fs.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

// Write saves a source that its owner may write. Everyone may read it.
// Nobody else may write or run it.
func TestWriteMakesAFileOthersCanReadAndNotChange(t *testing.T) {
	withoutAUmask(t)
	path := filepath.Join(t.TempDir(), "dusk.css")
	if err := terminal().Write(path); err != nil {
		t.Fatal(err)
	}
	if got := modeOf(t, path); got != aFileMode {
		t.Errorf("Write made a file of mode %o, wanted %o", got, aFileMode)
	}
}

// Rendered files get the same modes. The folders above them let everyone
// enter. Only the owner may write in them.
func TestWriteFileMakesItsFolderAndFileWithTheirModes(t *testing.T) {
	withoutAUmask(t)
	where := t.TempDir()
	path := filepath.Join(where, "a", "b", "out")
	if err := writeFile(path, writesWords); err != nil {
		t.Fatal(err)
	}
	if got := modeOf(t, path); got != aFileMode {
		t.Errorf("file mode %o, wanted %o", got, aFileMode)
	}
	if got := modeOf(t, filepath.Dir(path)); got != aFolderMode {
		t.Errorf("folder mode %o, wanted %o", got, aFolderMode)
	}
}

// The anti-test is a folder that cannot be made. The error is the folder's
// and not the file's.
func TestWriteFileSaysWhenItsFolderCannotBeMade(t *testing.T) {
	where := t.TempDir()
	block := filepath.Join(where, "block")
	if err := os.WriteFile(block, nil, aFileMode); err != nil {
		t.Fatal(err)
	}
	below := filepath.Join(block, "sub", "out")
	err := writeFile(below, writesNothing)
	var failed *fs.PathError
	if !errors.As(err, &failed) || failed.Op != "mkdir" {
		t.Errorf("the error was %v, wanted the folder's", err)
	}
}

// The text "/*/" opens a comment and does not close it. The anti-test is
// the same comment closed after.
func TestOpeningIsNotClosedByItsOwnOpening(t *testing.T) {
	for _, text := range []string{"/*/", "/*/ words", "/**", "/*"} {
		if got := Opening(text); got != "" {
			t.Errorf("Opening(%q) = %q", text, got)
		}
	}
	if got := Opening("/*/ words */"); got != "/ words" {
		t.Errorf("Opening of a comment that opens with a slash gave %q", got)
	}
	if got := Opening("/* words */"); got != "words" {
		t.Errorf("Opening gave %q", got)
	}
}

// Copy over a file records it as replaced. The anti-test is a copy where
// there was no file.
func TestCopyRecordsWhatItReplaced(t *testing.T) {
	where := t.TempDir()
	source := filepath.Join(where, "source")
	write(t, source, "new words\n")
	taken := filepath.Join(where, "taken")
	write(t, taken, "old words\n")
	free := filepath.Join(where, "free")
	var done Installed
	if err := done.Copy(source, taken, true); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(done.Replaced, []string{taken}) {
		t.Errorf("replaced %v, wanted %v", done.Replaced, taken)
	}
	if err := done.Copy(source, free, true); err != nil {
		t.Fatal(err)
	}
	if len(done.Replaced) != one || len(done.Written) != twoFiles {
		t.Errorf("replaced %v, written %v", done.Replaced, done.Written)
	}
}

// The disk is not asked where a link goes when the target is not a link.
// It is asked one look at the target and nothing else.
func TestCopyDoesNotFollowWhatIsNotALink(t *testing.T) {
	files := mocks.NewMockFiles(gomock.NewController(t))
	info := folderInfo(t)
	gomock.InOrder(
		files.EXPECT().ReadFile(theSource).Return(theWords, nil),
		files.EXPECT().Lstat(theTarget).Return(info, nil),
		files.EXPECT().ReadFile(theTarget).Return(theWords, nil))
	var done Installed
	if err := done.copy(files, theSource, theTarget, false); err != nil {
		t.Fatal(err)
	}
	if len(done.Same) != one {
		t.Errorf("the same file was recorded as %v", done)
	}
}

// Copy says a link to nothing is one. It goes no further. The test above
// is the anti-test.
func TestCopySaysALinkGoesToNothing(t *testing.T) {
	where := t.TempDir()
	link := filepath.Join(where, "link")
	if err := os.Symlink(filepath.Join(where, "nowhere"), link); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	files := mocks.NewMockFiles(gomock.NewController(t))
	gomock.InOrder(
		files.EXPECT().ReadFile(theSource).Return(theWords, nil),
		files.EXPECT().Lstat(theTarget).Return(info, nil),
		files.EXPECT().EvalSymlinks(theTarget).Return("", errRefusal))
	var done Installed
	got := done.copy(files, theSource, theTarget, true)
	if got == nil || got.Error() != theTarget+" is a link to nothing" {
		t.Errorf("a link to nothing gave %v", got)
	}
}

// Each ground the colourway leaves unset is the ground mixed toward the ink.
// Its share is a part of the distance between them in oklab.
func TestGroundsAreMixedByTheirShares(t *testing.T) {
	roles := css.New()
	roles.Set("ground", srgb.MustHex("#101020").Swatch())
	roles.Set("ink", srgb.MustHex("#d0d0e0").Swatch())
	resolved := Resolve(roles)
	ground, _ := roles.Get("ground")
	ink, _ := roles.Get("ink")
	whole := ok.Distance(ok.FromSwatch(ground), ok.FromSwatch(ink))
	shares := sharesPicked(t)
	if len(shares) == nothing {
		t.Fatal("no shares were picked")
	}
	for _, each := range shares {
		mixed, found := resolved.Of(each.name)
		if !found {
			t.Fatalf("%s was not made", each.name)
		}
		went := ok.Distance(ok.FromSwatch(ground), ok.FromSwatch(mixed))
		got := went / whole
		if got < each.share-closeEnough || got > each.share+closeEnough {
			t.Errorf("%s is %g of the way, wanted %g", each.name, got,
				each.share)
		}
	}
}

// grounded is a ground and the share it is mixed toward the ink.
type grounded struct {
	name  string
	share float64
}

// sharesPicked reads the shares of the three grounds from
// testdata/shares.json. The test stops if it cannot.
func sharesPicked(t *testing.T) []grounded {
	t.Helper()
	raw, err := os.ReadFile("testdata/shares.json")
	if err != nil {
		t.Fatal(err)
	}
	var numbers map[string]any
	if err := json.Unmarshal(raw, &numbers); err != nil {
		t.Fatal(err)
	}
	var shares []grounded
	for name, share := range numbers {
		if number, isNumber := share.(float64); isNumber {
			shares = append(shares, grounded{name, number})
		}
	}
	return shares
}

// Claude's band colours need a ground to mix from. With none they are left
// out. The anti-test is the same roles with a ground.
func TestAClaudeBandWithNoGroundIsLeftOut(t *testing.T) {
	roles := css.New()
	roles.Set("ok", srgb.MustHex("#55aa55").Swatch())
	if _, found := Resolve(roles).Of("claude-added"); found {
		t.Error("a band was made with no ground to mix from")
	}
	roles.Set("ground", srgb.MustHex("#101020").Swatch())
	if _, found := Resolve(roles).Of("claude-added"); !found {
		t.Error("a band was left out with a ground to mix from")
	}
}
