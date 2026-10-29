package main

import (
	"embed"
	"testing"
)

// These are the shell's statuses as numbers: 0 for worked and 1 for
// failed. 2 means the call made no sense. Tests that used main.go's names
// would pass for any value.
const (
	// shellSucceeded is the shell's status for a call that worked. It is
	// defined by the shell.
	shellSucceeded = 0
	// shellFailed is the shell's status for a call that could not work. It
	// is defined by the shell.
	shellFailed = 1
	// shellWrongCall is the status for a call that makes no sense. It is
	// defined by the shell.
	shellWrongCall = 2
)

// The three statuses are the shell's numbers, and three different ones.
func TestStatusesAreTheShells(t *testing.T) {
	asProcess(t, "")
	if got := run([]string{"swatch"}); got != shellSucceeded {
		t.Errorf("a good call stopped with %d", got)
	}
	if got := run([]string{"nosuch"}); got != shellWrongCall {
		t.Errorf("a wrong call stopped with %d", got)
	}
	t.Cleanup(sourcesWas{sources}.put)
	sources = embed.FS{}
	if got := run([]string{"swatch", "--go"}); got != shellFailed {
		t.Errorf("a failed call stopped with %d", got)
	}
}

// The anti-test: a good call is not the status of a failed or a wrong one.
func TestAGoodCallIsNotFailedOrWrong(t *testing.T) {
	asProcess(t, "")
	got := run([]string{"swatch", "--css"})
	if got == shellFailed || got == shellWrongCall {
		t.Errorf("a good call stopped with %d", got)
	}
}
