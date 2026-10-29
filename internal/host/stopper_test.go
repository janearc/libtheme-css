package host_test

import (
	"testing"

	"go.uber.org/mock/gomock"
	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/internal/host"
	"github.com/janearc/libtheme-css/internal/mocks"
)

// The numbers the stopper tests pick.
const (
	// wrongCall is a status that means the call made no sense. It is defined
	// by the shell's convention for the commands here.
	wrongCall = 2
	// widest is how far from zero a drawn status may fall. It is defined by
	// this test.
	widest = 1000
	// once is a call made one time. It is defined by counting.
	once = 1
)

// A function stopped through ExitFunc is called once, with the status.
func TestExitFuncCallsTheFunctionWithTheStatus(t *testing.T) {
	stopper := mocks.NewMockStopper(gomock.NewController(t))
	stopper.EXPECT().Exit(wrongCall).Times(once)
	var program host.Stopper = host.ExitFunc(stopper.Exit)
	program.Exit(wrongCall)
}

// The property: whatever the status, ExitFunc hands it on unchanged and
// calls the function once.
func TestExitFuncHandsOnEveryStatusUnchanged(t *testing.T) {
	rapid.Check(t, handsOnTheStatus)
}

// handsOnTheStatus is one draw of the property above.
func handsOnTheStatus(draw *rapid.T) {
	status := rapid.IntRange(-widest, widest).Draw(draw, "status")
	controller := gomock.NewController(draw)
	stopper := mocks.NewMockStopper(controller)
	stopper.EXPECT().Exit(status).Times(once)
	host.ExitFunc(stopper.Exit).Exit(status)
	controller.Finish()
}

// stoppedWithNothing is a call that stops with no function to stop with.
func stoppedWithNothing() { host.ExitFunc(nil).Exit(wrongCall) }

// panicked is whether the call panicked.
func panicked(call func()) (did bool) {
	defer note(&did)
	call()
	return did
}

// note is deferred to say that a panic happened.
func note(did *bool) { *did = recover() != nil }

// The anti-test: a stopper with no function does not quietly do nothing. A
// program that thought it had stopped would go on, so the call fails.
func TestExitFuncWithNoFunctionDoesNotQuietlySucceed(t *testing.T) {
	if !panicked(stoppedWithNothing) {
		t.Error("stopping with no function returned")
	}
}
