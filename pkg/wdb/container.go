// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Magic values that frame the sections of a database. Each one was found
// with `strings -a -t d` and is present in every corpus case.
const (
	fileMagic     = "Xilinx WAVE DATABASE 01"
	producerMagic = "Xilinx Simulator"
	typeMagic     = "Xilinx ISim TYPE FILE 001"
	debugMagic    = "Xilinx ISim DBG 006"

	dirEvent = "WDB.Event"
	dirRTTI  = "Xilinx RTTI"
	dirDBG   = "Xilinx DBG"
)

// Fixed offsets in the file header. See docs/format.md, "The container".
const (
	offDirPointers = 0x48 // three uint64 file offsets of directory entries
	offEvent       = 0xe0 // the WDB.Event region

	dirEntryLen  = 48 // [24 byte name][uint64 count][uint64 offset][uint64 length]
	dirNameLen   = 24
	pageDirStart = 0x30 // page records start this far into the Xilinx DBG entry
	pageDirLen   = 0x4c0
	pageDirComp  = 0x328 // offset of the compressed length inside a page record
	pageLen      = 10240 // every value page inflates to this many bytes
)

// DirEntry is one entry of the container directory. Three of them are
// reached from the pointers at offset 0x48, and each names one section.
type DirEntry struct {
	// Name is the NUL terminated ASCII name: WDB.Event, Xilinx RTTI or
	// Xilinx DBG.
	Name string
	// Count is 1 in every observed file.
	Count uint64
	// Offset is the file offset of the section the entry describes.
	Offset uint64
	// Length is the section's length in bytes.
	Length uint64
}

// Header is the fixed part of the file, up to the type table.
type Header struct {
	// Timestamp is the Unix time at which the database was written. It
	// differs between two runs of the same design, so it is noise.
	Timestamp uint32
	// EndTimePS is the simulation end time in picoseconds, from the
	// WDB.Event region.
	EndTimePS uint64
	// HasSignals is the flag at 0x110: 1 when any signal is logged.
	HasSignals bool
	// PageSize is the value at 0x120, the size a value page inflates to.
	PageSize uint32
	// Dirs holds the three directory entries in pointer order.
	Dirs []DirEntry
}

// PageRef is one entry of the page directory that follows the Xilinx DBG
// directory entry. Each one locates a compressed value page.
type PageRef struct {
	// Offset is the file offset of the zlib stream.
	Offset uint64
	// CompressedLen is the length of that stream in bytes.
	CompressedLen uint64
}

func cstring(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

// readHeader decodes the fixed header and the directory.
func readHeader(d []byte) (Header, error) {
	var h Header
	if len(d) < 0x128 {
		return h, fmt.Errorf("file is %d bytes, shorter than the fixed header", len(d))
	}
	if cstring(d[:dirNameLen]) != fileMagic {
		return h, fmt.Errorf("no %q magic at offset 0", fileMagic)
	}
	if cstring(d[0x18:0x30]) != producerMagic {
		return h, fmt.Errorf("no %q producer name at offset 0x18", producerMagic)
	}
	h.Timestamp = binary.LittleEndian.Uint32(d[0x38:])
	h.EndTimePS = binary.LittleEndian.Uint64(d[offEvent:])
	h.HasSignals = binary.LittleEndian.Uint32(d[0x110:]) != 0
	h.PageSize = binary.LittleEndian.Uint32(d[0x120:])
	for i := 0; i < 3; i++ {
		p := binary.LittleEndian.Uint64(d[offDirPointers+8*i:])
		if p+dirEntryLen > uint64(len(d)) {
			return h, fmt.Errorf("directory pointer %d at %#x points past the file", i, p)
		}
		e := d[p : p+dirEntryLen]
		h.Dirs = append(h.Dirs, DirEntry{
			Name:   cstring(e[:dirNameLen]),
			Count:  binary.LittleEndian.Uint64(e[24:]),
			Offset: binary.LittleEndian.Uint64(e[32:]),
			Length: binary.LittleEndian.Uint64(e[40:]),
		})
	}
	return h, nil
}

// Dir returns the directory entry with the given name.
func (h *Header) Dir(name string) (DirEntry, bool) {
	for _, e := range h.Dirs {
		if e.Name == name {
			return e, true
		}
	}
	return DirEntry{}, false
}

// readPageDir decodes the page directory. It starts pageDirStart bytes
// into the Xilinx DBG directory entry, which itself sits at the end of the
// section it describes. The directory is a run of pageDirLen byte records
// and a 16 byte trailer, and the first page follows the trailer directly,
// so the first page offset fixes the record count.
func readPageDir(d []byte, dbg DirEntry) ([]PageRef, error) {
	p := dbg.Offset + dbg.Length + pageDirStart
	if p > uint64(len(d)) {
		return nil, fmt.Errorf("page directory at %#x runs past the file", p)
	}
	if p+pageDirLen > uint64(len(d)) {
		// A database with no value pages ends here, with no trailer.
		return nil, nil
	}
	first := binary.LittleEndian.Uint64(d[p+8:])
	if first < p+16 || first > uint64(len(d)) {
		return nil, fmt.Errorf("first page offset %#x is not after the page directory at %#x", first, p)
	}
	span := first - p - 16
	if span%pageDirLen != 0 {
		return nil, fmt.Errorf("page directory span %d is not a multiple of %d", span, pageDirLen)
	}
	n := int(span / pageDirLen)
	refs := make([]PageRef, 0, n)
	for i := 0; i < n; i++ {
		r := d[p+uint64(i)*pageDirLen:]
		off := binary.LittleEndian.Uint64(r[8:])
		clen := binary.LittleEndian.Uint64(r[pageDirComp:])
		if off+clen > uint64(len(d)) {
			return nil, fmt.Errorf("page %d at %#x+%d runs past the file", i, off, clen)
		}
		refs = append(refs, PageRef{Offset: off, CompressedLen: clen})
	}
	return refs, nil
}
