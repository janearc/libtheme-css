// Package smoke holds the tests that load generated colour schemes in a real
// vim and neovim. A file compared only as text can still be wrong as code,
// so these tests start the editors and ask what the file left behind.
//
// The editors run headless and without any config of the user's. There is
// nothing here but tests.
//
// A machine without the editor fails the test that needs it, because a
// check that cannot run does not pass.
package smoke
