// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// Record is one value change inside a page.
type Record struct {
	// TimePS is the simulation time of the change in picoseconds.
	TimePS uint64
	// Key names the object within the page; see Object.Key.
	Key uint32
	// Data is the value, encoded as docs/format.md describes per kind.
	Data []byte
}

// Page is one inflated value page.
type Page struct {
	// Ref locates the page's zlib stream in the file.
	Ref PageRef
	// T0 and T1 are the two times the page header holds. T0 has been 0
	// in every observed page; T1 has been the end time.
	T0, T1  uint64
	Records []Record
}

// readPage inflates one page and splits it into records.
func readPage(d []byte, ref PageRef) (Page, error) {
	pg := Page{Ref: ref}
	zr, err := zlib.NewReader(bytes.NewReader(d[ref.Offset : ref.Offset+ref.CompressedLen]))
	if err != nil {
		return pg, fmt.Errorf("page at %#x: %w", ref.Offset, err)
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return pg, fmt.Errorf("page at %#x: inflate: %w", ref.Offset, err)
	}
	if len(raw) != pageLen {
		return pg, fmt.Errorf("page at %#x inflates to %d bytes, want %d", ref.Offset, len(raw), pageLen)
	}
	c := &cursor{b: raw}
	pg.T0 = c.u64()
	pg.T1 = c.u64()
	n := int(c.u32())
	for i := 0; i < n; i++ {
		r := Record{TimePS: c.u64(), Key: c.u32()}
		l := int(c.u32())
		if !c.need(l) {
			break
		}
		r.Data = raw[c.p : c.p+l]
		c.p += l
		pg.Records = append(pg.Records, r)
	}
	if c.err != nil {
		return pg, fmt.Errorf("page at %#x: record %d: %w", ref.Offset, len(pg.Records), c.err)
	}
	for _, b := range raw[c.p:] {
		if b != 0 {
			return pg, fmt.Errorf("page at %#x has non-zero padding after %d records", ref.Offset, n)
		}
	}
	return pg, nil
}

// Change is one value of one object at one time.
type Change struct {
	TimePS uint64
	Data   []byte
}

// Changes returns every recorded value of the object, in page order,
// which is time order in every observed file.
func (f *File) Changes(o Object) ([]Change, error) {
	p := o.Page()
	if p >= len(f.Pages) {
		return nil, fmt.Errorf("object handle %#x names page %d of %d", o.Handle, p, len(f.Pages))
	}
	var out []Change
	for _, r := range f.Pages[p].Records {
		if r.Key == o.Key() {
			out = append(out, Change{TimePS: r.TimePS, Data: r.Data})
		}
	}
	return out, nil
}
