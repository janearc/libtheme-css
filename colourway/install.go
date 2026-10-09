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

// places says where each program reads its rendered files. The key is the
// folder the file sits in under a colourways folder. Most places are in the
// user's config folder. The vim place is in the home folder.
var places = map[string]struct {
	inConfig bool
	path     string
}{
	"ghostty":     {true, "ghostty/themes"},
	"nvim/colors": {true, "nvim/colors"},
	"vim/colors":  {false, ".vim/colors"},
	"glamour":     {true, "glamour"},
}

// shipped are the folders an install copies: sources and their outputs.
var shipped = []string{"sources", "ghostty", "nvim/colors", "vim/colors",
	"glamour"}

// Config is the user's config folder. It is XDG_CONFIG_HOME if that is set
// to an absolute path, as the spec asks, and ~/.config otherwise.
func Config(home string) string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(home, ".config")
}

// Place returns where a program reads the files in a folder of a
// colourways folder. It returns false if the folder is not a program's.
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

// Folder is the user's colourways folder. It is colourways or colorways in
// the config folder, whichever exists. If neither exists, it is colorways
// when LANG starts with en_US, and colourways otherwise.
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

// Installed records what an install did with each file. Written files were
// written, and Same files were already the same. Kept files are the user's
// files that differed and were left alone, since replace was not asked for.
// Replaced files took the place of a user's file that differed.
type Installed struct {
	Written, Same, Kept, Replaced []string
}

// Install copies the sources and renders of a colourways folder into the
// user's own colourways folder. It also copies each render to where its
// program reads it. A user's file that differs is kept unless replace is
// true, so nothing they tuned is lost.
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

// Copy puts one file in place and records which it was: written, the same,
// or kept. If the target is a link, Copy writes through to the file the link
// names. A link to nothing is an error. So is a file that cannot be read.
// Neither is overwritten.
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

// writeWhole writes a temporary file beside the target and renames it into
// place. A write cut short then leaves the old file whole, not half of the
// new one.
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

// Say tells the reader what an install did, under the command they ran. It
// says where the colourways went and how many files were written, already
// the same, and kept. It names each file it replaced and each file of
// theirs it kept. If any were kept, it says how to replace them.
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
