package colourway

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/janearc/libtheme-css/internal/mocks"
)

// shippedFolder is a colourways folder with one source and its renders
// for each program.
func shippedFolder(t *testing.T) string {
	t.Helper()
	from := t.TempDir()
	for path, text := range map[string]string{
		"sources/dusk.css":     ":root {\n  --ground: #101020;\n}\n",
		"ghostty/dusk":         "background = #101020\n",
		"nvim/colors/dusk.lua": "-- dusk\n",
		"vim/colors/dusk.vim":  "\" dusk\n",
		"glamour/dusk.json":    "{}\n",
	} {
		write(t, filepath.Join(from, path), text)
	}
	return from
}

// write puts text in a file, making its folder.
func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestInstallPutsEachFileWhereItIsRead installs into an empty home and
// finds the sources in the user's colourways folder and each render where
// its program reads it.
func TestInstallPutsEachFileWhereItIsRead(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("LANG", "en_GB.UTF-8")
	done, err := Install(shippedFolder(t), home, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		".config/colourways/sources/dusk.css",
		".config/colourways/vim/colors/dusk.vim",
		".config/ghostty/themes/dusk",
		".config/nvim/colors/dusk.lua",
		".vim/colors/dusk.vim",
		".config/glamour/dusk.json",
	} {
		if _, err := os.Stat(filepath.Join(home, path)); err != nil {
			t.Errorf("no %s", path)
		}
	}
	if len(done.Written) != 9 || len(done.Kept) != 0 {
		t.Errorf("wrote %d and kept %d", len(done.Written), len(done.Kept))
	}
}

// TestInstallKeepsWhatTheUserTuned finds a theme of the user's that
// differs kept, then replaced only when replacing is asked for.
func TestInstallKeepsWhatTheUserTuned(t *testing.T) {
	home, from := t.TempDir(), shippedFolder(t)
	t.Setenv("XDG_CONFIG_HOME", "")
	theirs := filepath.Join(home, ".config/ghostty/themes/dusk")
	write(t, theirs, "background = #202030\n")
	done, err := Install(from, home, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(done.Kept) != 1 || done.Kept[0] != theirs {
		t.Errorf("kept %v", done.Kept)
	}
	if raw, _ := os.ReadFile(theirs); string(raw) != "background = #202030\n" {
		t.Error("a theme the user tuned was overwritten")
	}
	again, err := Install(from, home, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Written) != 0 || len(again.Same) != 8 {
		t.Errorf("a second install wrote %d and found %d the same",
			len(again.Written), len(again.Same))
	}
	if _, err := Install(from, home, true); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(theirs); string(raw) != "background = #101020\n" {
		t.Error("replacing did not replace")
	}
}

// TestFolderSpelling finds the folder a user already has, and a new one
// spelled for the locale.
func TestFolderSpelling(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("LANG", "en_US.UTF-8")
	if got := Folder(home); filepath.Base(got) != "colorways" {
		t.Errorf("an American locale with neither folder gets %s", got)
	}
	t.Setenv("LANG", "nl_NL.UTF-8")
	if got := Folder(home); filepath.Base(got) != "colourways" {
		t.Errorf("a Dutch locale with neither folder gets %s", got)
	}
	write(t, filepath.Join(home, ".config/colorways/sources/x.css"), "")
	if got := Folder(home); filepath.Base(got) != "colorways" {
		t.Errorf("a user with a colorways folder gets %s", got)
	}
}

// TestInstallFollowsTheConfigFolder puts each program's file in the
// config folder XDG_CONFIG_HOME names, where the program reads it, and
// vim's in the home.
func TestInstallFollowsTheConfigFolder(t *testing.T) {
	home, config := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)
	if _, err := Install(shippedFolder(t), home, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(config, "ghostty/themes/dusk"),
		filepath.Join(config, "nvim/colors/dusk.lua"),
		filepath.Join(home, ".vim/colors/dusk.vim")} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("no %s", path)
		}
	}
}

// TestInstallWritesThroughALink finds a theme that is a link to a file
// elsewhere written at the file it names, the link left a link; and a
// link to nothing refused.
func TestInstallWritesThroughALink(t *testing.T) {
	home, elsewhere := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	real := filepath.Join(elsewhere, "dusk")
	write(t, real, "background = #202030\n")
	link := filepath.Join(home, ".config/ghostty/themes/dusk")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(shippedFolder(t), home, true); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Error("the link was replaced by a file")
	}
	if raw, _ := os.ReadFile(real); string(raw) != "background = #101020\n" {
		t.Error("the file the link names was not written")
	}
	if err := os.Remove(real); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(shippedFolder(t), home, true); err == nil {
		t.Error("a link to nothing was written through")
	}
}

// TestInstallRefusesWhatItCannotRead stops at a file of the user's it
// cannot read, rather than writing over it.
func TestInstallRefusesWhatItCannotRead(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	theirs := filepath.Join(home, ".config/ghostty/themes/dusk")
	write(t, theirs, "background = #202030\n")
	if err := os.Chmod(theirs, 0o200); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(shippedFolder(t), home, false); err == nil {
		t.Error("a file that could not be read was not an error")
	}
	if err := os.Chmod(theirs, 0o644); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(theirs); string(raw) != "background = #202030\n" {
		t.Error("a file that could not be read was written over")
	}
}

// TestInstallWantsAColourwaysFolder refuses a folder with no sources,
// rather than saying it wrote nothing and succeeding.
func TestInstallWantsAColourwaysFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	if _, err := Install(filepath.Join(t.TempDir(), "nothing"), home, false); err == nil {
		t.Error("a folder with no sources installed without a word")
	}
}

// Say tells what an install did: where, how many, each file replaced and
// each of theirs kept; and, if any were kept, how to replace them with a
// command that exists.
func TestSayTellsWhatAnInstallDid(t *testing.T) {
	var done Installed
	done.Written = []string{"a", "b"}
	done.Same = []string{"c"}
	done.Kept = []string{"d"}
	done.Replaced = []string{"b"}
	var out strings.Builder
	done.Say(&out, t.TempDir())
	text := out.String()
	for _, want := range []string{"written 2, already the same 1, kept 1",
		"replaced yours, which differed: b", "kept yours, which differs: d",
		"paratune -install -replace"} {
		if !strings.Contains(text, want) {
			t.Errorf("no %q in\n%s", want, text)
		}
	}
	if strings.Contains(text, "colourway install") {
		t.Errorf("Say advertises a command colourway does not have:\n%s", text)
	}
}

// Nothing kept, nothing to say about replacing.
func TestSayOffersNoReplaceWhenNothingWasKept(t *testing.T) {
	var done Installed
	done.Written = []string{"a"}
	var out strings.Builder
	done.Say(&out, t.TempDir())
	if strings.Contains(out.String(), "replace") {
		t.Errorf("Say offered to replace what was not kept:\n%s", out.String())
	}
}

// The numbers the install tests below pick. Each is defined by the comment
// on it.
const (
	// one is a single file or error. It is defined by counting.
	one = 1
	// folderMode is the mode of a folder an install makes. It is defined by
	// unix, and by install.go, which asks for it.
	folderMode fs.FileMode = 0o755
	// fileMode is the mode of a file an install writes. It is defined by
	// unix, and by install.go, which asks for it.
	fileMode fs.FileMode = 0o644
)

// The names the mock install is given. Each is defined by this test.
const (
	// theSource is a file in the colourways folder being installed.
	theSource = "from/ghostty/dusk"
	// theTarget is where the source is to go.
	theTarget = "to/ghostty/themes/dusk"
	// theTemporary is the name the temporary file is made with.
	theTemporary = "to/ghostty/themes/.install-1234"
	// aThemeFile is the pattern a temporary file is made with. It is defined
	// by install.go.
	aThemeFile = ".install-*"
)

// errRefusal is the error a mocked file system gives. It is not a missing
// file.
var errRefusal = errors.New("the disk refused")

// theWords is what the source holds.
var theWords = []byte("background = #101020\n")

// A colourways folder with only sources in it installs the sources and goes
// on: a folder it lacks is not an error.
func TestInstallGoesOnWhenAFolderIsMissing(t *testing.T) {
	home, from := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	write(t, filepath.Join(from, "sources/dusk.css"), ":root {\n}\n")
	done, err := Install(from, home, false)
	if err != nil || len(done.Written) != one || len(done.Kept) != nothing {
		t.Errorf("wrote %v, kept %v: %v", done.Written, done.Kept, err)
	}
}

// The anti-test: a hidden file and a folder among the renders are not
// installed. Only files that are not hidden are.
func TestInstallLeavesHiddenFilesAndFoldersBehind(t *testing.T) {
	home, from := t.TempDir(), shippedFolder(t)
	t.Setenv("XDG_CONFIG_HOME", "")
	write(t, filepath.Join(from, "ghostty/.hidden"), "x\n")
	write(t, filepath.Join(from, "ghostty/sub/inner"), "x\n")
	done, err := Install(from, home, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append(done.Written, done.Same...) {
		if strings.Contains(path, ".hidden") || strings.Contains(path, "sub") {
			t.Errorf("installed %s", path)
		}
	}
}

// The anti-test: a copy whose source cannot be read is an error, and it
// records nothing.
func TestCopyRefusesASourceItCannotRead(t *testing.T) {
	var done Installed
	where := t.TempDir()
	missing, target := filepath.Join(where, "nowhere"), filepath.Join(where, "target")
	err := done.Copy(missing, target, false)
	if err == nil || len(done.Written) != nothing {
		t.Errorf("a missing source: %v, wrote %v", err, done.Written)
	}
}

// folderInfo is the file info of a folder that exists.
func folderInfo(t *testing.T) fs.FileInfo {
	t.Helper()
	info, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return info
}

// The anti-test: a colourways folder that cannot be listed stops the install
// with the disk's error, and nothing is read or written after it.
func TestInstallStopsWhenAFolderCannotBeListed(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	files := mocks.NewMockFiles(gomock.NewController(t))
	folder := folderInfo(t)
	gomock.InOrder(
		files.EXPECT().Stat("from/sources").Return(folder, nil),
		files.EXPECT().ReadDir("from/sources").Return(nil, errRefusal))
	done, err := install(files, "from", home, false)
	if !errors.Is(err, errRefusal) || len(done.Written) != nothing {
		t.Errorf("wrote %v: %v", done.Written, err)
	}
}

// fault is the error a step gets when it is the one failing, and nothing
// when it is not.
func fault(step, failing int) error {
	if step == failing {
		return errRefusal
	}
	return nil
}

// The steps of writing a file whole, in order. Each is defined by
// writeWhole.
const (
	// makeFolder is the step that makes the target's folder. It is defined
	// by writeWhole, as the first.
	makeFolder = iota
	// makeTemporary is the step that makes the temporary file beside the
	// target. It is defined by writeWhole.
	makeTemporary
	// writeTemporary is the step that writes the bytes. It is defined by
	// writeWhole.
	writeTemporary
	// changeMode is the step that sets the file's mode. It is defined by
	// writeWhole.
	changeMode
	// closeTemporary is the step that closes the file. It is defined by
	// writeWhole.
	closeTemporary
	// renameTemporary is the step that moves the file into place. It is
	// defined by writeWhole, as the last.
	renameTemporary
	// allWell is no step failing. It is defined by this test.
	allWell
)

// expectWrite says what a mocked disk is asked, in order, while a file is
// written whole and the step named by failing fails. A step past the failing
// one is not asked. Once the temporary file exists it is removed at the end,
// and closed if the failure came before the close.
func expectWrite(files *mocks.MockFiles, temporary *mocks.MockFile,
	failing int) {
	folder, broken := filepath.Dir(theTarget), fault(makeFolder, failing)
	calls := []any{files.EXPECT().ReadFile(theSource).Return(theWords, nil),
		files.EXPECT().Lstat(theTarget).Return(nil, fs.ErrNotExist),
		files.EXPECT().ReadFile(theTarget).Return(nil, fs.ErrNotExist),
		files.EXPECT().MkdirAll(folder, folderMode).Return(broken)}
	if failing > makeFolder {
		broken := fault(makeTemporary, failing)
		calls = append(calls, files.EXPECT().CreateTemp(folder,
			aThemeFile).Return(temporary, broken))
	}
	if failing > makeTemporary {
		written, broken := len(theWords), fault(writeTemporary, failing)
		calls = append(calls, temporary.EXPECT().Name().Return(theTemporary),
			temporary.EXPECT().Write(theWords).Return(written, broken))
	}
	if failing == writeTemporary {
		calls = append(calls, temporary.EXPECT().Close().Return(nil))
	}
	if failing > writeTemporary {
		broken := fault(changeMode, failing)
		calls = append(calls, temporary.EXPECT().Chmod(fileMode).Return(broken))
	}
	if failing == changeMode {
		calls = append(calls, temporary.EXPECT().Close().Return(nil))
	}
	if failing > changeMode {
		broken := fault(closeTemporary, failing)
		calls = append(calls, temporary.EXPECT().Close().Return(broken))
	}
	if failing > closeTemporary {
		broken := fault(renameTemporary, failing)
		calls = append(calls, files.EXPECT().Rename(theTemporary,
			theTarget).Return(broken))
	}
	if failing > makeTemporary {
		calls = append(calls, files.EXPECT().Remove(theTemporary).Return(nil))
	}
	gomock.InOrder(calls...)
}

// The anti-test: whichever step of writing a file fails, the failing disk's
// error comes back, the install records nothing as written, a temporary
// file that was made is removed, and no step after the failing one is tried.
// When none fails the file is recorded once.
func TestAFaultAtAnyStepLeavesNothingRecorded(t *testing.T) {
	for failing := makeFolder; failing <= allWell; failing++ {
		controller := gomock.NewController(t)
		files := mocks.NewMockFiles(controller)
		expectWrite(files, mocks.NewMockFile(controller), failing)
		var done Installed
		err := done.copy(files, theSource, theTarget, false)
		if failing == allWell {
			if err != nil || len(done.Written) != one {
				t.Errorf("no fault: %v, wrote %v", err, done.Written)
			}
			continue
		}
		if !errors.Is(err, errRefusal) || len(done.Written) != nothing {
			t.Errorf("fault at step %d: %v, wrote %v", failing, err,
				done.Written)
		}
	}
}
