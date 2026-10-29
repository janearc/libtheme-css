package main

import (
	"io"
	"os"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/janearc/libtheme-css/internal/host"
	"github.com/janearc/libtheme-css/internal/mocks"
)

// process is the program main runs in, for a test. It has a mock to stop at,
// the files its output went to, and the words typed to it.
type process struct {
	stopper         *mocks.MockStopper
	out, errs, keys *os.File
}

// was is what a test changed about the program, and puts it back.
type was struct {
	arguments       []string
	out, errs, keys *os.File
	stopsWith       host.Stopper
	made            process
}

// restore puts the program back as it was.
func (before was) restore() {
	os.Args, os.Stdout, os.Stderr, os.Stdin, program = before.arguments,
		before.out, before.errs, before.keys, before.stopsWith
	before.made.out.Close()
	before.made.errs.Close()
	before.made.keys.Close()
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

// startOfFile is the offset of a file's first byte. It is defined by io.
const startOfFile = 0

// typed is a new file holding the words a person typed, read from the start.
func typed(t *testing.T, words string) *os.File {
	t.Helper()
	file := emptyFile(t)
	if _, err := file.WriteString(words); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(startOfFile, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	return file
}

// asProcess makes the program a test can call main in. It sets the arguments,
// sends the output to files, types the words, and holds the exit with a mock.
func asProcess(t *testing.T, words string, arguments ...string) process {
	t.Helper()
	made := process{
		stopper: mocks.NewMockStopper(gomock.NewController(t)),
		out:     emptyFile(t),
		errs:    emptyFile(t),
		keys:    typed(t, words),
	}
	before := was{os.Args, os.Stdout, os.Stderr, os.Stdin, program, made}
	t.Cleanup(before.restore)
	os.Args = append([]string{"visualdocs"}, arguments...)
	os.Stdout, os.Stderr, os.Stdin = made.out, made.errs, made.keys
	program = made.stopper
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
