// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"fmt"
	"io"
	"strings"
)

// Dump writes every decoded structure of the file as text, one section
// per structure, in file order. It is the reproduction command behind
// the findings in docs/format.md: every row there can be checked by
// reading the matching line of this output.
func (f *File) Dump(w io.Writer) error {
	p := func(format string, args ...any) {
		fmt.Fprintf(w, format, args...)
	}
	h := &f.Header
	p("header:\n")
	p("  timestamp      %d\n", h.Timestamp)
	p("  end time       %d ps\n", h.EndTimePS)
	p("  has signals    %v\n", h.HasSignals)
	p("  marker         at %#x, %d\n", h.MarkerOffset, h.Marker)
	p("  page size      %d\n", h.PageSize)
	p("  arena table    %d slots:", len(h.ArenaOffsets))
	for _, a := range h.ArenaOffsets {
		p(" %#x", a)
	}
	p("\n")
	p("directory:\n")
	for _, e := range h.Dirs {
		p("  %-12q count %d  offset %#x  length %d\n", e.Name, e.Count, e.Offset, e.Length)
	}

	p("types (%d):\n", len(f.Types))
	for i, t := range f.Types {
		p("  [%d] %s %s", i, t.Kind, t.Name)
		switch t.Kind {
		case KindEnum:
			p(" class %d: %s", t.Class, strings.Join(t.Literals, " "))
		case KindInteger:
			p(" %d to %d", t.Low, t.High)
		case KindReal:
			p(" %g to %g", t.FLow, t.FHigh)
		case KindPhysical:
			for _, u := range t.Units {
				p(" %s=%d", u.Name, u.Scale)
			}
		case KindArray:
			p(" of [%d] indexed by [%d], %d dim", t.Elem, t.Index, t.Dims)
			if t.Constrained {
				p(", constrained")
			}
			for _, r := range t.Ranges {
				p(" (%s)", r)
			}
		case KindRecord:
			for _, fd := range t.Fields {
				p(" %s:[%d]", fd.Name, fd.Type)
				for _, r := range fd.Ranges {
					p("(%s)", r)
				}
			}
		}
		p("\n")
	}

	d := &f.Debug
	p("debug:\n")
	p("  timestamp      %d\n", d.Timestamp)
	p("  counts         scopes %d, units %d, objects %d, declarations %d\n",
		d.Counts[0], d.Counts[1], d.Counts[2], d.Counts[3])
	p("  offsets       ")
	for _, o := range d.Offsets {
		p(" %#x", o)
	}
	p("\n")

	p("files (%d):\n", len(f.Files))
	for i, sf := range f.Files {
		if sf.Path == "" && sf.Local == "" {
			continue
		}
		p("  [%d] %s", i, sf.Path)
		if sf.Local != sf.Path {
			p(" (local %s)", sf.Local)
		}
		p("\n")
	}

	p("units (%d):\n", len(f.Units))
	for i, u := range f.Units {
		p("  [%d] %s", i, u.Kind)
		if u.Name != "" {
			p(" %s(%s)", u.Name, u.Arch)
		}
		p(" decls %d+%d file %d line %d entity file %d line %d\n",
			u.FirstDecl, u.NumDecls, u.File, u.Line, u.EntityFile, u.EntityLine)
	}

	p("scopes (%d):\n", len(f.Scopes))
	for i, s := range f.Scopes {
		p("  [%d] %-24s parent %d children %d+%d first object %d unit %d file %d line %d\n",
			i, f.Path(i), s.Parent, s.FirstChild, s.NumChildren, s.FirstObject, s.Unit, s.File, s.Line)
	}

	p("declarations (%d):\n", len(f.Decls))
	for i, dc := range f.Decls {
		p("  [%d] %s %s : [%d] %s, %d bytes, file %d line %d", i, dc.Kind, dc.Name, dc.Type, f.Types[dc.Type].Name, dc.Size, dc.File, dc.Line)
		for _, r := range dc.Ranges {
			p(" (%s)", r)
		}
		p("\n")
	}

	p("objects (%d):\n", len(f.Objects))
	for i, o := range f.Objects {
		p("  [%d] handle %#x arena %d key %#x  %s", i, o.Handle, o.Arena(), o.Key(), f.ObjectPath(o))
		if o.Generic {
			p(" (no second handle)")
		}
		p("\n")
	}

	p("arenas (%d):\n", len(f.Arenas))
	for i, a := range f.Arenas {
		p("  [%d] record at %#x, %d pages\n", i, a.Offset, len(a.Pages))
		for j, pg := range f.Pages[i] {
			p("    page %d at %#x, %d bytes compressed, t0 %d t1 %d, %d records\n",
				j, pg.Ref.Offset, pg.Ref.CompressedLen, pg.T0, pg.T1, len(pg.Records))
			for _, r := range pg.Records {
				p("      t=%-10d key %#05x  % x\n", r.TimePS, r.Key, r.Data)
			}
		}
	}

	p("values:\n")
	for _, o := range f.Objects {
		dc := f.Decls[o.Decl]
		ch, err := f.Changes(o)
		if err != nil {
			return err
		}
		for _, c := range ch {
			v, err := f.Decode(dc, c.Data)
			if err != nil {
				return fmt.Errorf("%s at %d ps: %w", f.ObjectPath(o), c.TimePS, err)
			}
			p("  t=%-10d %s = %s\n", c.TimePS, f.ObjectPath(o), f.String(v))
		}
	}
	return nil
}
