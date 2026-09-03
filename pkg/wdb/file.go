// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"fmt"
	"os"
	"strings"
)

// File is a decoded waveform database. Every field is a direct decoding
// of bytes in the file; Value and Path derive from them.
//
// The layout was worked out from the corpus under //hdl/corpus, on files
// written by Vivado 2025.2 alone. docs/format.md records each field and
// the comparison that found it.
type File struct {
	Header Header
	Debug  Debug

	// Types is the type table, in file order. Decl.Type and the array
	// and record entries index into it.
	Types []Type
	// Scopes is the instance tree in preorder, starting at `_top`.
	Scopes []Scope
	// Units are the elaborated design units.
	Units []Unit
	// Decls are the declared signals and generics, per unit.
	Decls []Decl
	// Files is the source file table.
	Files []SourceFile
	// Objects are the instantiated signals and generics, one per scope
	// that declares them.
	Objects []Object
	// Arenas is the page directory. Pages holds each arena's pages,
	// inflated, in the same order.
	Arenas []Arena
	Pages  [][]Page
}

// Read decodes a whole database held in memory.
func Read(d []byte) (*File, error) {
	f := &File{}
	var err error
	if f.Header, err = readHeader(d); err != nil {
		return nil, err
	}
	rtti, ok := f.Header.Dir(dirRTTI)
	if !ok {
		return nil, fmt.Errorf("no %q directory entry", dirRTTI)
	}
	dbg, ok := f.Header.Dir(dirDBG)
	if !ok {
		return nil, fmt.Errorf("no %q directory entry", dirDBG)
	}
	for _, e := range []DirEntry{rtti, dbg} {
		if e.Offset+e.Length > uint64(len(d)) {
			return nil, fmt.Errorf("section %q at %#x+%d runs past the file", e.Name, e.Offset, e.Length)
		}
	}
	if f.Types, err = readTypes(d, rtti); err != nil {
		return nil, err
	}
	if err = readDebug(f, d, dbg); err != nil {
		return nil, err
	}
	if f.Arenas, err = readPageDir(d, &f.Header, dbg); err != nil {
		return nil, err
	}
	for _, a := range f.Arenas {
		var pages []Page
		for _, ref := range a.Pages {
			pg, err := readPage(d, ref)
			if err != nil {
				return nil, err
			}
			pages = append(pages, pg)
		}
		f.Pages = append(f.Pages, pages)
	}
	// The logged ranges name the objects that have records. Every
	// object inside a range has at least one; no object outside has
	// any. t6_var_int has an unlogged variable after the last range,
	// t9_mark_gap an unlogged package constant before the first, and
	// t9_mark_two both.
	for _, r := range f.Header.Logged {
		if r[1] >= uint64(len(f.Objects)) {
			return nil, fmt.Errorf("logged range [%d, %d] names an object past the %d objects", r[0], r[1], len(f.Objects))
		}
		for i := r[0]; i <= r[1]; i++ {
			f.Objects[i].Logged = true
		}
	}
	for i, o := range f.Objects {
		ch, err := f.Changes(o)
		if err != nil {
			return nil, fmt.Errorf("object %d (%s): %w", i, f.ObjectPath(o), err)
		}
		if (len(ch) > 0) != o.Logged {
			return nil, fmt.Errorf("object %d (%s) has %d records, the logged ranges say logged=%v", i, f.ObjectPath(o), len(ch), o.Logged)
		}
	}
	return f, nil
}

// ReadFile decodes the database at path.
func ReadFile(path string) (*File, error) {
	d, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := Read(d)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return f, nil
}

// Path is the dotted instance path of a scope, without the `_top` root:
// `tb`, `tb.dut`, `tb.dut.p`.
func (f *File) Path(scope int) string {
	var parts []string
	for s := scope; s > 0 && s < len(f.Scopes); s = f.Scopes[s].Parent {
		parts = append(parts, f.Scopes[s].Name)
	}
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}

// ObjectPath is the dotted path of an object: its scope's path and its
// declared name.
func (f *File) ObjectPath(o Object) string {
	return f.Path(o.Scope) + "." + f.Decls[o.Decl].Name
}
