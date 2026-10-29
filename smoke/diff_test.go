package smoke_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// duskSheet is a source with the colours every program needs. The ink is left
// to the test, and the heading is its own so markdown has something to say.
const duskSheet = `:root {
  --ground: #202020;
  --ink: %s;
  --heading: #aa5555;
  --black: #101010; --red: #aa5555; --green: #55aa55; --yellow: #aaaa55;
  --blue: #5555aa; --magenta: #aa55aa; --cyan: #55aaaa; --white: #aaaaaa;
  --bright-black: #555555; --bright-red: #ff5555; --bright-green: #55ff55;
  --bright-yellow: #ffff55; --bright-blue: #5555ff;
  --bright-magenta: #ff55ff; --bright-cyan: #55ffff; --bright-white: #ffffff;
}
`

// kinds are the files a render makes, and the tracked name each would have
// in a repository.
var kinds = []struct{ rendered, tracked string }{
	{"ghostty/dusk", "ghostty/dusk"},
	{"nvim/colors/dusk.lua", "dusk.lua"},
	{"glamour/dusk.json", "dusk.json"},
	{"sources/dusk.css", "dusk.css"},
}

// dirMode is the permission a folder is made with. It is defined here.
const dirMode = os.FileMode(0o755)

// pair is a file before and after.
type pair struct {
	before, after string
}

// built builds the colourway command into a folder of its own and returns
// its path. These tests run the binary, as git does.
func built(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "colourway")
	build := exec.Command("go", "build", "-o", binary,
		"github.com/janearc/libtheme-css/cmd/colourway")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}
	return binary
}

// rendered renders the source with the ink and returns the folder, which
// holds the source as sources/dusk.css and every program's file.
func rendered(t *testing.T, binary, ink string) string {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(dir, "sources", "dusk.css")
	if err := os.MkdirAll(filepath.Dir(source), dirMode); err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(duskSheet, "%s", ink)
	if err := os.WriteFile(source, []byte(text), private); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(binary, "render", source, dir).CombinedOutput(); err != nil {
		t.Fatalf("render failed: %v\n%s", err, output)
	}
	return dir
}

// renderedPairs renders two inks apart and returns each file's pair by its
// tracked name.
func renderedPairs(t *testing.T, binary string) map[string]pair {
	t.Helper()
	before, after := rendered(t, binary, "#e0e0e0"), rendered(t, binary, "#f0f0f0")
	pairs := map[string]pair{}
	for _, kind := range kinds {
		pairs[kind.tracked] = pair{filepath.Join(before, kind.rendered),
			filepath.Join(after, kind.rendered)}
	}
	return pairs
}

// bare copies a file to a name with no extension, as a driver that does
// not keep the tracked name would hand it.
func bare(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	copied := filepath.Join(t.TempDir(), "XXXXXX")
	if err := os.WriteFile(copied, raw, private); err != nil {
		t.Fatal(err)
	}
	return copied
}

// drive runs the command as git's driver with the arguments, and returns
// what it printed.
func drive(t *testing.T, binary string, arguments ...string) string {
	t.Helper()
	output, err := exec.Command(binary, append([]string{"diff", "--git"}, arguments...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", arguments, err, output)
	}
	return string(output)
}

// somethingChanged is what a diff says when it saw a role change.
var somethingChanged = regexp.MustCompile(`, [1-9][0-9]* changed`)

// someAdded and someRemoved are what a diff says when one side has roles the
// other lacks.
var (
	someAdded   = regexp.MustCompile(`, [1-9][0-9]* added`)
	someRemoved = regexp.MustCompile(`, [1-9][0-9]* removed`)
)

// nothingSeen is what a diff says when it read no roles on either side.
const nothingSeen = "0 the same, 0 changed, 0 added, 0 removed"

// Git hands the driver the tracked name, then each side as a temp file.
// The dialect is the tracked name's, so temp files with no extension still
// read as the css, lua and json they are.
func TestDriverReadsTheTrackedNameNotTheTempName(t *testing.T) {
	binary := built(t)
	for tracked, files := range renderedPairs(t, binary) {
		output := drive(t, binary, tracked, bare(t, files.before), "abc",
			"100644", bare(t, files.after), "def", "100644")
		if !somethingChanged.MatchString(output) {
			t.Errorf("%s: the driver saw no change:\n%s", tracked, output)
		}
	}
}

// The same for a rename, which gives each side its own tracked name. The same
// for an added and a removed file, whose other side is /dev/null. Each says
// what it saw. Renames show a change. Added files show roles added. Removed
// files show roles removed. Seeing nothing is not enough to pass.
func TestDriverReadsRenamesAndAdditions(t *testing.T) {
	binary := built(t)
	files := renderedPairs(t, binary)["dusk.css"]
	old, fresh := bare(t, files.before), bare(t, files.after)
	for name, seen := range map[string]struct {
		arguments []string
		want      *regexp.Regexp
	}{
		"rename": {[]string{"old.css", old, "abc", "100644", fresh, "def",
			"100644", "new.css", "similarity index 90%"}, somethingChanged},
		"added": {[]string{"new.css", os.DevNull, "000", "100644", fresh,
			"def", "100644"}, someAdded},
		"removed": {[]string{"old.css", old, "abc", "100644", os.DevNull,
			"000", "100644"}, someRemoved},
	} {
		output := drive(t, binary, seen.arguments...)
		if strings.Contains(output, nothingSeen) || !seen.want.MatchString(output) {
			t.Errorf("%s: the driver did not say what it saw:\n%s", name, output)
		}
	}
}

// The anti-test: a file that is not the dialect its tracked name says is
// said so, not reported as a diff of nothing.
func TestDriverSaysWhenAFileIsNotWhatItsNameSays(t *testing.T) {
	binary := built(t)
	files := renderedPairs(t, binary)["dusk.lua"]
	output := drive(t, binary, "dusk.css", files.before, "abc", "100644",
		files.after, "def", "100644")
	if strings.Contains(output, nothingSeen) || !strings.Contains(output, "no colours") {
		t.Errorf("lua read as css was not called out:\n%s", output)
	}
}

// git runs git in a repository with none of the user's config, and fails
// the test if git fails.
func git(t *testing.T, repository string, arguments ...string) string {
	t.Helper()
	run := exec.Command("git", append([]string{"-C", repository}, arguments...)...)
	run.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null", "GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com")
	output, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", arguments, err, output)
	}
	return string(output)
}

// version is one commit of the test repository: each file's text.
type version map[string]string

// committed writes the versions into the repository, one commit each.
func committed(t *testing.T, repository string, versions []version) {
	t.Helper()
	git(t, repository, "init", "-q", ".")
	for _, files := range versions {
		for name, text := range files {
			path := filepath.Join(repository, name)
			if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(text), private); err != nil {
				t.Fatal(err)
			}
		}
		git(t, repository, "add", "-A")
		git(t, repository, "commit", "-q", "--no-gpg-sign", "-m", "version")
	}
}

// The smoke test: real git, a repository with the driver set for a source
// and a terminal theme, and one role changed in each. Git calls the built
// command, which says what changed in both.
func TestGitCallsTheDriver(t *testing.T) {
	binary := built(t)
	repository := t.TempDir()
	attributes := "*.css diff=colourway\nghostty/* diff=colourway\n"
	committed(t, repository, []version{
		{".gitattributes": attributes, "dusk.css": ":root {\n  --ink: #ffffff;\n}\n",
			"ghostty/dusk": "foreground = #ffffff\n"},
		{"dusk.css": ":root {\n  --ink: #eeeeee;\n}\n",
			"ghostty/dusk": "foreground = #eeeeee\n"},
	})
	shown := git(t, repository, "-c", "diff.colourway.command="+binary+" diff --git",
		"diff", "HEAD~1", "HEAD")
	for _, name := range []string{"dusk.css", "ghostty/dusk"} {
		if !strings.Contains(shown, "colourway diff "+name+"\n") {
			t.Errorf("git did not call the driver for %s:\n%s", name, shown)
		}
	}
	if !somethingChanged.MatchString(shown) {
		t.Errorf("the driver saw no change:\n%s", shown)
	}
}
