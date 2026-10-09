// Package glamour is the glamour dialect: a markdown style said from a
// colourway's roles, and read back into them.
//
// Glamour, which glow and others draw markdown with, takes a style as
// json: for each part of a document, its colour, its ground and how it
// is set. Code blocks go to chroma, whose colours sit in the same file.
//
// This dialect sets every colour in the style from a role, and leaves
// how things are set (margins, prefixes, bold) to base.json, a style
// that reads well.
//
// The roles are md-something. Each falls back to the roles a colourway
// for a terminal and an editor already has, and in the end to the ink,
// so a colourway with no markdown roles still makes a style that agrees
// with its vim: headings in heading, links in link, comments in dim.
//
// A ground with nothing to fall back to is left out, and the part sits
// on the page. No terminal paints a code block's ground: chroma clears
// it and glamour does not draw it. md-block-ground is written for the
// renderers that do.
package glamour

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

//go:embed base.json
var base []byte

// chroma is where chroma's token colours sit in a style.
const chroma = "code_block.chroma."

// Role is one of the style's colours: the role a colourway names it
// by, the roles it falls back to in order, and the keys in the style it
// sets.
type Role struct {
	Name      string
	Fallbacks []string
	Keys      []string
}

// Roles are the style's colours. The grounds have no last resort; every
// other role ends at the ink.
var Roles = []Role{
	{"md-body", list("ink"), list("document.color")},
	{"md-heading", list("heading"), list("heading.color")},
	{"md-heading-ground", nil, list("heading.background_color")},
	{"md-small-heading", list("link"),
		list("h3.color", "h4.color", "h5.color", "h6.color")},
	{"md-small-heading-ground", nil, list("h3.background_color",
		"h4.background_color", "h5.background_color",
		"h6.background_color")},
	{"md-emphasis", list("md-body"), list("emph.color")},
	{"md-strong", list("md-body"), list("strong.color")},
	{"md-link", list("link"), list("link.color")},
	{"md-link-text", list("md-link", "link"),
		list("link_text.color", "image.color")},
	{"md-image-text", list("dim"), list("image_text.color")},
	{"md-rule", list("border"), list("hr.color")},
	{"md-bullet", list("md-body"),
		list("item.color", "enumeration.color")},
	{"md-quote", list("dim"), list("block_quote.color")},
	{"md-code", list("code"), list("code.color")},
	{"md-block-text", list("md-code", "code"),
		list("code_block.color", chroma+"text.color")},
	{"md-code-ground", list("panel"), list("code.background_color")},
	{"md-block-ground", list("md-code-ground", "panel"),
		list(chroma + "background.background_color")},
	{"md-keyword", list("heading"), chromaKeys("keyword",
		"keyword_reserved", "keyword_namespace", "name_tag",
		"name_exception", "literal_string_escape")},
	{"md-type", list("code"), chromaKeys("keyword_type",
		"name_builtin", "generic_inserted")},
	{"md-function", list("link"), chromaKeys("name_function",
		"name_class", "name_attribute")},
	{"md-string", list("string"), chromaKeys("literal_string",
		"literal", "literal_date", "name_decorator")},
	{"md-number", list("string"),
		chromaKeys("literal_number", "name_constant")},
	{"md-comment", list("dim"),
		chromaKeys("comment", "comment_preproc")},
	{"md-operator", list("ink"), chromaKeys("operator", "punctuation")},
	{"md-name", list("md-code", "code"),
		chromaKeys("name", "name_other")},
	{"md-error", list("err"), chromaKeys("error", "generic_deleted")},
	{"md-subheading", list("md-small-heading", "link"),
		chromaKeys("generic_subheading")},
}

// list is names written as a list.
func list(names ...string) []string {
	return names
}

// chromaKeys are the colour keys of kinds of chroma's tokens.
func chromaKeys(kinds ...string) []string {
	keys := make([]string, len(kinds))
	for position, kind := range kinds {
		keys[position] = chroma + kind + ".color"
	}
	return keys
}

// ground is whether a role is a ground, which has no last resort.
func ground(role string) bool {
	return strings.HasSuffix(role, "-ground")
}

// Of is the style a colourway's roles describe, as glamour reads it.
// The ink is required, or md-body in its place, since every colour that
// is not a ground ends there.
func Of(roles *css.Sheet) ([]byte, error) {
	if _, found := lastResort(roles); !found {
		return nil, fmt.Errorf("glamour: no ink, and no md-body")
	}
	style := map[string]any{}
	if err := json.Unmarshal(base, &style); err != nil {
		return nil, fmt.Errorf("glamour: base.json: %w", err)
	}
	for _, entry := range Roles {
		value, found := resolve(roles, entry.Name, entry.Fallbacks)
		for _, key := range entry.Keys {
			if !found {
				remove(style, key)
				continue
			}
			set(style, key, hex(value))
		}
	}
	return json.MarshalIndent(style, "", "  ")
}

// resolve is a role's colour: its own, else its fallbacks' in order,
// each of those followed through its own fallbacks, and only then the
// last resort, unless it is a ground.
func resolve(roles *css.Sheet, role string,
	fallbacks []string) (swatch.Swatch, bool) {
	if value, found := follow(roles, role, fallbacks); found {
		return value, true
	}
	if ground(role) {
		return swatch.Swatch{}, false
	}
	return lastResort(roles)
}

// follow is a role's colour from itself and its fallbacks alone.
func follow(roles *css.Sheet, role string,
	fallbacks []string) (swatch.Swatch, bool) {
	for _, name := range append([]string{role}, fallbacks...) {
		if value, found := roles.Get(name); found {
			return value, true
		}
		for _, entry := range Roles {
			if entry.Name != name || name == role {
				continue
			}
			value, found := follow(roles, name, entry.Fallbacks)
			if found {
				return value, true
			}
		}
	}
	return swatch.Swatch{}, false
}

// lastResort is the colour every role that is not a ground ends at: the
// ink, or md-body when the roles are a style's, read back.
func lastResort(roles *css.Sheet) (swatch.Swatch, bool) {
	if value, found := roles.Get("ink"); found {
		return value, true
	}
	return roles.Get("md-body")
}

// Read is a style's colours as a colourway's roles: each role from the
// first of its keys the style sets. A key that is not a colour is an
// error naming it.
func Read(source []byte) (*css.Sheet, error) {
	style := map[string]any{}
	if err := json.Unmarshal(source, &style); err != nil {
		return nil, fmt.Errorf("glamour: %w", err)
	}
	roles := css.New()
	for _, entry := range Roles {
		for _, key := range entry.Keys {
			text, found := lookup(style, key)
			if !found {
				continue
			}
			lamps, err := srgb.FromHex(text)
			if err != nil {
				return nil, fmt.Errorf("glamour: %s: %w",
					key, err)
			}
			roles.Set(entry.Name, lamps.Swatch())
			break
		}
	}
	return roles, nil
}

// lookup is the string at a dotted key, as "heading.color".
func lookup(style map[string]any, key string) (string, bool) {
	var node any = style
	for _, part := range strings.Split(key, ".") {
		object, found := node.(map[string]any)
		if !found {
			return "", false
		}
		node = object[part]
	}
	text, found := node.(string)
	return text, found
}

// set writes a value at a dotted key, making the objects on the way.
func set(style map[string]any, key, value string) {
	parts := strings.Split(key, ".")
	node := style
	for _, part := range parts[:len(parts)-1] {
		next, found := node[part].(map[string]any)
		if !found {
			next = map[string]any{}
			node[part] = next
		}
		node = next
	}
	node[parts[len(parts)-1]] = value
}

// remove takes the value at a dotted key out, if it is there.
func remove(style map[string]any, key string) {
	parts := strings.Split(key, ".")
	node := style
	for _, part := range parts[:len(parts)-1] {
		next, found := node[part].(map[string]any)
		if !found {
			return
		}
		node = next
	}
	delete(node, parts[len(parts)-1])
}

// hex is a swatch as the #rrggbb a style holds.
func hex(value swatch.Swatch) string {
	lamps, _ := srgb.FromSwatch(value)
	return lamps.Hex()
}
