package colourway

import (
	"slices"
	"strconv"
	"strings"

	"github.com/janearc/libtheme-css/dialects/claude"
	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/dialects/nvim"
)

// knownRoles are the names a dialect or the general vocabulary reads as a
// role. The numbered ones, the markdown ones and the nvim- overrides are
// not among them.
func knownRoles() []string {
	names := slices.Concat(Grounds, Text, States, ghostty.Ansi[:])
	for _, key := range ghostty.Keys {
		names = append(names, key.Role)
	}
	for _, slot := range claude.Slots {
		names = append(names, slot.Role)
	}
	for _, entry := range nvim.Roles {
		names = append(names, entry.Role)
	}
	return names
}

// numbered is whether a name is palette-N for a whole number N, which
// ghostty and claude code read.
func numbered(name string) bool {
	digits, found := strings.CutPrefix(name, "palette-")
	if !found {
		return false
	}
	_, err := strconv.Atoi(digits)
	return err == nil
}

// markdown is whether a name is one of glamour's markdown roles, which all
// begin md-.
func markdown(name string) bool {
	return strings.HasPrefix(name, "md-")
}

// isRole is whether a dialect reads a custom property of this name as a
// role. It is a known name, a palette number or a markdown role. It may be
// a known name behind nvim- instead.
func isRole(name string) bool {
	known := knownRoles()
	if slices.Contains(known, name) || numbered(name) || markdown(name) {
		return true
	}
	overridden, found := strings.CutPrefix(name, "nvim-")
	return found && slices.Contains(known, overridden)
}
