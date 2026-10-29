package mocks_test

import (
	"errors"
	"io/fs"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/janearc/libtheme-css/internal/host"
	"github.com/janearc/libtheme-css/internal/mocks"
)

const (
	// aStatus is a status to stop with. It is defined by this test.
	aStatus = 3
	// aMode is a mode to ask for. It is defined by this test.
	aMode fs.FileMode = 0o640
	// aCount is the bytes a write reports. It is defined by this test.
	aCount = 7
)

// errRefused is the error the mocks give back when told to.
var errRefused = errors.New("refused")

// The mock of a Stopper is a Stopper, and holds a call to what it was told to
// expect: the status it was given, once.
func TestTheStopperMockIsAStopper(t *testing.T) {
	mock := mocks.NewMockStopper(gomock.NewController(t))
	var stopper host.Stopper = mock
	mock.EXPECT().Exit(aStatus)
	stopper.Exit(aStatus)
}

// The mock of Files is Files: each call it is told to expect gives back what
// it was told to give, the error among them.
func TestTheFilesMockGivesWhatItIsTold(t *testing.T) {
	mock := mocks.NewMockFiles(gomock.NewController(t))
	var files host.Files = mock
	mock.EXPECT().Stat("a").Return(nil, errRefused)
	mock.EXPECT().Lstat("b").Return(nil, errRefused)
	mock.EXPECT().ReadDir("c").Return(nil, errRefused)
	mock.EXPECT().ReadFile("d").Return([]byte("d"), nil)
	mock.EXPECT().EvalSymlinks("e").Return("f", nil)
	mock.EXPECT().MkdirAll("g", aMode).Return(errRefused)
	mock.EXPECT().CreateTemp("h", "i").Return(nil, errRefused)
	mock.EXPECT().Remove("j").Return(errRefused)
	mock.EXPECT().Rename("k", "l").Return(errRefused)
	_, errStat := files.Stat("a")
	_, errLstat := files.Lstat("b")
	_, errReadDir := files.ReadDir("c")
	raw, errRead := files.ReadFile("d")
	named, errEval := files.EvalSymlinks("e")
	_, errTemp := files.CreateTemp("h", "i")
	for name, err := range map[string]error{"stat": errStat, "lstat": errLstat,
		"readdir": errReadDir, "mkdirall": files.MkdirAll("g", aMode),
		"createtemp": errTemp, "remove": files.Remove("j"),
		"rename": files.Rename("k", "l")} {
		if !errors.Is(err, errRefused) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if string(raw) != "d" || errRead != nil || named != "f" || errEval != nil {
		t.Errorf("read %q (%v), evaluated %q (%v)", raw, errRead, named, errEval)
	}
}

// The mock of File is a File. Its name, write, mode and close each give what
// they were told.
func TestTheFileMockGivesWhatItIsTold(t *testing.T) {
	mock := mocks.NewMockFile(gomock.NewController(t))
	var file host.File = mock
	mock.EXPECT().Name().Return("a")
	mock.EXPECT().Write([]byte("b")).Return(aCount, nil)
	mock.EXPECT().Chmod(aMode).Return(errRefused)
	mock.EXPECT().Close().Return(errRefused)
	name := file.Name()
	written, err := file.Write([]byte("b"))
	if name != "a" || written != aCount || err != nil {
		t.Errorf("name %q, wrote %d (%v)", name, written, err)
	}
	if !errors.Is(file.Chmod(aMode), errRefused) ||
		!errors.Is(file.Close(), errRefused) {
		t.Error("chmod or close did not give the error it was told to")
	}
}
