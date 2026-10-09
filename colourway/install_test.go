package colourway

import (
	"os"
	"path/filepath"
	"testing"
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
