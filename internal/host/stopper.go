// Package host is what a command asks of the machine it runs on: the way it
// stops, and the files it reads and writes. Each is an interface, so a test
// can hand in a mock and make the machine fail where a real one seldom does.
package host

// Stopper stops the program with a status.
type Stopper interface {
	Exit(status int)
}

// ExitFunc is a function that stops the program, such as os.Exit, as a
// Stopper. A command holds os.Exit as one, and a test holds a mock in its
// place, so main can run and say what it stopped with.
type ExitFunc func(status int)

// Exit stops the program by calling the function.
func (stop ExitFunc) Exit(status int) { stop(status) }
