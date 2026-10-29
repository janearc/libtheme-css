package host_test

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/internal/host"
)

// The modes the tests use. Each is defined by unix.
const (
	// folderMode is a folder its owner can use and others can read. It is
	// defined by unix.
	folderMode fs.FileMode = 0o755
	// fileMode is a file its owner can write and others can read. It is
	// defined by unix.
	fileMode fs.FileMode = 0o644
)

// How many entries a listing has, and where its first is.
const (
	// one is a single entry in a folder. It is defined by counting.
	one = 1
	// first is where the first entry of a listing is. It is defined by the
	// list.
	first = 0
)

// words is what the tests write into a file.
var words = []byte("a colour is a colour wherever it lands\n")

// A file made through the machine's file system is on disk. It is read back
// whole, found by Stat and Lstat, listed, and gone once removed.
func TestDiskIsTheMachinesFiles(t *testing.T) {
	disk := host.NewDisk()
	deep := filepath.Join(t.TempDir(), "a", "b")
	if err := disk.MkdirAll(deep, folderMode); err != nil {
		t.Fatal(err)
	}
	made, err := disk.CreateTemp(deep, ".install-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := made.Write(words); err != nil {
		t.Fatal(err)
	}
	if err := made.Chmod(fileMode); err != nil {
		t.Fatal(err)
	}
	if err := made.Close(); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(deep, "theme")
	if err := disk.Rename(made.Name(), target); err != nil {
		t.Fatal(err)
	}
	got, err := disk.ReadFile(target)
	if err != nil || !bytes.Equal(got, words) {
		t.Errorf("read %q, %v", got, err)
	}
	if info, err := disk.Stat(target); err != nil || info.IsDir() {
		t.Errorf("stat: %v, %v", info, err)
	}
	if info, err := disk.Lstat(target); err != nil || info.IsDir() {
		t.Errorf("lstat: %v, %v", info, err)
	}
	listed, err := disk.ReadDir(deep)
	if err != nil || len(listed) != one || listed[first].Name() != "theme" {
		t.Errorf("listed %v, %v", listed, err)
	}
	if err := disk.Remove(target); err != nil {
		t.Fatal(err)
	}
	if _, err := disk.Stat(target); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("after remove: %v", err)
	}
}

// EvalSymlinks follows a link to the file it names. Lstat sees the link.
func TestDiskFollowsALink(t *testing.T) {
	disk := host.NewDisk()
	where := t.TempDir()
	actual := filepath.Join(where, "real")
	link := filepath.Join(where, "link")
	if err := os.WriteFile(actual, words, fileMode); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(actual, link); err != nil {
		t.Fatal(err)
	}
	named, err := disk.EvalSymlinks(link)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := disk.EvalSymlinks(actual); named != want {
		t.Errorf("the link names %s, not %s", named, want)
	}
	if info, err := disk.Lstat(link); err != nil ||
		info.Mode().Type() != fs.ModeSymlink {
		t.Errorf("lstat of a link: %v, %v", info, err)
	}
}

// The anti-test: every call on a path that is not there is refused with the
// platform's error. A file that could not be made comes back as nil, not as a
// nil file wrapped in the interface.
func TestDiskRefusesWhatIsNotThere(t *testing.T) {
	disk := host.NewDisk()
	nowhere := filepath.Join(t.TempDir(), "nowhere")
	inside := filepath.Join(nowhere, "inside")
	refusals := map[string]error{}
	_, refusals["stat"] = disk.Stat(nowhere)
	_, refusals["lstat"] = disk.Lstat(nowhere)
	_, refusals["readdir"] = disk.ReadDir(nowhere)
	_, refusals["readfile"] = disk.ReadFile(nowhere)
	_, refusals["evalsymlinks"] = disk.EvalSymlinks(nowhere)
	refusals["remove"] = disk.Remove(nowhere)
	refusals["rename"] = disk.Rename(nowhere, inside)
	made, err := disk.CreateTemp(nowhere, ".install-*")
	refusals["createtemp"] = err
	for call, err := range refusals {
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s of a missing path: %v", call, err)
		}
	}
	if made != nil {
		t.Errorf("a file that could not be made came back as %v", made)
	}
}

// The anti-test: a folder cannot be made where a file is.
func TestDiskCannotMakeAFolderOverAFile(t *testing.T) {
	disk := host.NewDisk()
	occupied := filepath.Join(t.TempDir(), "occupied")
	if err := os.WriteFile(occupied, words, fileMode); err != nil {
		t.Fatal(err)
	}
	if err := disk.MkdirAll(filepath.Join(occupied, "under"),
		folderMode); err == nil {
		t.Error("a folder was made inside a file")
	}
}

// The file system NewDisk gives is the machine's. It reads a file that is
// there, and refuses one that is not.
func TestNewDiskIsTheMachinesFiles(t *testing.T) {
	where := filepath.Join(t.TempDir(), "there")
	if err := os.WriteFile(where, words, fileMode); err != nil {
		t.Fatal(err)
	}
	files := host.NewDisk()
	if got, err := files.ReadFile(where); err != nil || !bytes.Equal(got, words) {
		t.Errorf("read %q, %v", got, err)
	}
	if _, err := files.ReadFile(where + "-not"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a file that is not there gave %v", err)
	}
}

// The property: whatever bytes are written through a temporary file and
// renamed into place are the bytes read back, and the temporary name is
// gone.
func TestWhatIsWrittenIsWhatIsRead(t *testing.T) {
	rapid.Check(t, folder{t.TempDir()}.writesWhatIsRead)
}

// folder is a place on disk a property writes in.
type folder struct{ path string }

// writesWhatIsRead is one draw of the property above.
func (in folder) writesWhatIsRead(draw *rapid.T) {
	disk := host.NewDisk()
	raw := rapid.SliceOf(rapid.Byte()).Draw(draw, "raw")
	made, err := disk.CreateTemp(in.path, ".install-*")
	if err != nil {
		draw.Fatal(err)
	}
	if _, err := made.Write(raw); err != nil {
		draw.Fatal(err)
	}
	if err := made.Close(); err != nil {
		draw.Fatal(err)
	}
	temporary := made.Name()
	target := filepath.Join(in.path, "written")
	if err := disk.Rename(temporary, target); err != nil {
		draw.Fatal(err)
	}
	got, err := disk.ReadFile(target)
	if err != nil || !bytes.Equal(got, raw) {
		draw.Fatalf("wrote %d bytes and read %d: %v", len(raw), len(got), err)
	}
	if _, err := disk.Stat(temporary); !errors.Is(err, fs.ErrNotExist) {
		draw.Fatalf("the temporary file is still there: %v", err)
	}
}
