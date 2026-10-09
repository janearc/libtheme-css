package colourway

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// places are where each program reads a rendered file, by the folder the
// file sits in under a colourways folder: in the user's config folder,
// or, for vim, in the home itself.
var places = map[string]struct {
	inConfig bool
	path     string
}{
	"ghostty":     {true, "ghostty/themes"},
	"nvim/colors": {true, "nvim/colors"},
	"vim/colors":  {false, ".vim/colors"},
	"glamour":     {true, "glamour"},
}

// shipped are the folders of a colourways folder an install takes: the
// sources, and what they render into.
var shipped = []string{"sources", "ghostty", "nvim/colors", "vim/colors",
	"glamour"}

// Config is the user's config folder: XDG_CONFIG_HOME when it is set to an
// absolute path, as the spec asks, and ~/.config otherwise.
func Config(home string) string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(home, ".config")
}

// Place is where a program reads the files rendered into a folder of a
// colourways folder, and whether the folder is a program's at all.
func Place(folder, home string) (string, bool) {
	place, known := places[folder]
	if !known {
		return "", false
	}
	if place.inConfig {
		return filepath.Join(Config(home), place.path), true
	}
	return filepath.Join(home, place.path), true
}

// Folder is where a user's colourways live: colourways or colorways in
// their config folder, whichever is there; if neither is, colorways for
// an American English locale and colourways for everyone else.
func Folder(home string) string {
	for _, spelling := range []string{"colourways", "colorways"} {
		dir := filepath.Join(Config(home), spelling)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	if strings.HasPrefix(os.Getenv("LANG"), "en_US") {
		return filepath.Join(Config(home), "colorways")
	}
	return filepath.Join(Config(home), "colourways")
}

// Installed is what an install did with each file it was given: written,
// already the same, or kept, because a file of the user's was there and
// differed, and replacing was not asked for. Replaced are the written
// files that took the place of one of the user's that differed.
type Installed struct {
	Written, Same, Kept, Replaced []string
}

// Install copies a colourways folder's sources and renders into the
// user's own colourways folder, where paratune looks, and each render to
// where its program reads it. A file of theirs that differs is kept
// unless replace is asked for, so nothing they tuned is lost.
func Install(from, home string, replace bool) (Installed, error) {
	done, into := Installed{}, Folder(home)
	if info, err := os.Stat(filepath.Join(from, "sources")); err != nil ||
		!info.IsDir() {
		return done, fmt.Errorf("%s has no sources/ to install: "+
			"name a colourways folder, such as colourways/ in a "+
			"clone of chroma", from)
	}
	for _, folder := range shipped {
		entries, err := os.ReadDir(filepath.Join(from, folder))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return done, err
		}
		for _, entry := range entries {
			source := filepath.Join(from, folder, entry.Name())
			info, err := os.Stat(source)
			if err != nil || info.IsDir() ||
				strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			name := entry.Name()
			targets := []string{filepath.Join(into, folder, name)}
			if place, known := Place(folder, home); known {
				targets = append(targets,
					filepath.Join(place, name))
			}
			for _, target := range targets {
				err := done.Copy(source, target, replace)
				if err != nil {
					return done, err
				}
			}
		}
	}
	return done, nil
}

// Copy puts one file in place and says which it was: written, the same,
// or kept. A link is written through, to the file it names; a link to
// nothing, or a file that cannot be read, is an error, never overwritten.
func (done *Installed) Copy(source, target string, replace bool) error {
	raw, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if linked, err := os.Lstat(target); err == nil &&
		linked.Mode()&fs.ModeSymlink != 0 {
		named, err := filepath.EvalSymlinks(target)
		if err != nil {
			return fmt.Errorf("%s is a link to nothing", target)
		}
		target = named
	}
	there, err := os.ReadFile(target)
	existed := err == nil
	switch {
	case existed && bytes.Equal(there, raw):
		done.Same = append(done.Same, target)
		return nil
	case existed && !replace:
		done.Kept = append(done.Kept, target)
		return nil
	case !existed && !errors.Is(err, fs.ErrNotExist):
		return err
	}
	if err := writeWhole(target, raw); err != nil {
		return err
	}
	done.Written = append(done.Written, target)
	if existed {
		done.Replaced = append(done.Replaced, target)
	}
	return nil
}

// writeWhole writes a file beside its place and renames it in, so a write
// cut short leaves the old file whole, never half of the new one.
func writeWhole(target string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".install-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(raw); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), target)
}

// Say tells a reader what an install did, under the command they ran:
// where the colourways went, how many files of each kind, and every file
// of theirs it kept, by name.
func (done Installed) Say(out io.Writer, home string) {
	fmt.Fprintf(out, "colourways in %s, and each program's files "+
		"where it reads them\n", Folder(home))
	fmt.Fprintf(out, "written %d, already the same %d, kept %d\n",
		len(done.Written), len(done.Same), len(done.Kept))
	for _, path := range done.Replaced {
		fmt.Fprintln(out, "replaced yours, which differed:", path)
	}
	for _, path := range done.Kept {
		fmt.Fprintln(out, "kept yours, which differs:", path)
	}
	if len(done.Kept) > 0 {
		fmt.Fprintln(out, "install again asking to replace "+
			"(colourway install --replace, or paratune -install "+
			"with -replace) to take ours in their place")
	}
}
