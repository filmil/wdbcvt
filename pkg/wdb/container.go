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
	offDirPointers = 0x48  // three uint64 file offsets of directory entries
	offArenas      = 0xc8  // first slot of the arena table
	trailerLen     = 0x48  // the fixed words that end the header
	arenaSpan      = 0x800 // bytes of handle space per arena

	dirEntryLen  = 48 // [24 byte name][uint64 count][uint64 offset][uint64 length]
	dirNameLen   = 24
	pageDirStart = dirEntryLen // arena records follow the Xilinx DBG entry
	arenaRecLen  = 0x4c0       // one arena record
	arenaOffsets = 0x8         // uint64 page offsets, one per page
	arenaLens    = 0x328       // uint32 compressed lengths, one per page
	arenaCount   = 0x4b8       // uint64 number of pages
	arenaMax     = 100         // pages one arena record can name
	pageLen      = 10240       // every value page inflates to this many bytes
	chunkLen     = 146         // a value wider than this is written as chunks; see Changes
	markerLen    = 16
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
//
// The header is fixed up to offset 0xc8. From there a table of uint64
// arena record offsets runs to the trailer, and the trailer's position is
// known from the first directory pointer, which names the WDB.Event entry
// that follows it. The table has one slot per 0x800 bytes of handle
// space, rounded up: three slots for a testbench of one signal, four for
// seven to twelve one bit signals, six for twenty.
type Header struct {
	// Timestamp is the Unix time at which the database was written. It
	// differs between two runs of the same design, so it is noise.
	Timestamp uint32
	// ArenaOffsets are the slots of the arena table. Slot i is the file
	// offset of arena record i, or 0 when there is no such arena.
	ArenaOffsets []uint64
	// EndTimePS is the simulation end time in picoseconds, the first
	// word of the trailer.
	EndTimePS uint64
	// HandleSpace is the trailer word at 0x18 from its start: the number
	// of bytes of handle space the objects occupy. Each one bit signal
	// costs 0xf0 of it, and the arena table has ceil(HandleSpace/0x800)
	// slots. The trailer word at 0x0c repeats that slot count.
	HandleSpace uint64
	// MarkerOffset is the trailer word at 0x118 in a three slot header:
	// the file offset of the logged object ranges, or 0 when no object
	// is logged. It sits between the last arena record and the first
	// page unless a page was flushed before the simulation ended, and
	// then it follows that page.
	MarkerOffset uint64
	// Logged lists the ranges of object indices that have records, as
	// [first, last] pairs, read from MarkerOffset. The trailer word at
	// 0x110 counts them. One range [0, n-1] was first read as a single
	// marker word; t9_port_rec, with an unlogged package constant
	// between two logged objects, has two.
	Logged [][2]uint64
	// PageSize is the trailer word at 0x120, the size a value page
	// inflates to.
	PageSize uint32
	// Dirs holds the three directory entries in pointer order.
	Dirs []DirEntry
}

// PageRef locates one compressed value page.
type PageRef struct {
	// Offset is the file offset of the zlib stream.
	Offset uint64
	// CompressedLen is the length of that stream in bytes.
	CompressedLen uint64
}

// Arena is one record of the page directory. Objects are numbered by
// handle, and handle>>11 selects the arena whose pages hold the object's
// records. An arena starts a new page when the current one is full, so a
// record names one page per 10240 bytes of records.
type Arena struct {
	// Offset is the file offset of the record, 0 for an arena that was
	// never written.
	Offset uint64
	// Pages are the arena's pages in the order they were written, which
	// is time order. More than arenaMax of them means the record chained
	// to one or more continuation records.
	Pages []PageRef
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
	// The first directory pointer names the entry right after the
	// trailer, so it fixes where the arena table ends.
	trailer := binary.LittleEndian.Uint64(d[offDirPointers:]) - trailerLen
	if trailer < offArenas || (trailer-offArenas)%8 != 0 || trailer+trailerLen > uint64(len(d)) {
		return h, fmt.Errorf("header trailer at %#x does not follow the arena table at %#x", trailer, offArenas)
	}
	for p := uint64(offArenas); p < trailer; p += 8 {
		h.ArenaOffsets = append(h.ArenaOffsets, binary.LittleEndian.Uint64(d[p:]))
	}
	t := d[trailer:]
	h.EndTimePS = binary.LittleEndian.Uint64(t[0x00:])
	if n := binary.LittleEndian.Uint32(t[0x0c:]); int(n) != len(h.ArenaOffsets) {
		return h, fmt.Errorf("trailer says %d arena table slots, the table has %d", n, len(h.ArenaOffsets))
	}
	h.HandleSpace = binary.LittleEndian.Uint64(t[0x18:])
	if want := (h.HandleSpace + arenaSpan - 1) / arenaSpan; uint64(len(h.ArenaOffsets)) != want {
		return h, fmt.Errorf("arena table has %d slots, handle space %#x wants %d", len(h.ArenaOffsets), h.HandleSpace, want)
	}
	ranges := binary.LittleEndian.Uint64(t[0x30:])
	h.MarkerOffset = binary.LittleEndian.Uint64(t[0x38:])
	h.PageSize = binary.LittleEndian.Uint32(t[0x40:])
	if (h.MarkerOffset == 0) != (ranges == 0) {
		return h, fmt.Errorf("marker at %#x with %d logged ranges", h.MarkerOffset, ranges)
	}
	if h.MarkerOffset+ranges*markerLen > uint64(len(d)) {
		return h, fmt.Errorf("%d logged ranges at %#x run past the file", ranges, h.MarkerOffset)
	}
	for i := uint64(0); i < ranges; i++ {
		m := d[h.MarkerOffset+i*markerLen:]
		r := [2]uint64{binary.LittleEndian.Uint64(m), binary.LittleEndian.Uint64(m[8:])}
		if r[0] > r[1] {
			return h, fmt.Errorf("logged range %d is [%d, %d]", i, r[0], r[1])
		}
		if i > 0 && r[0] <= h.Logged[i-1][1] {
			return h, fmt.Errorf("logged range %d starts at %d, after the previous ended at %d", i, r[0], h.Logged[i-1][1])
		}
		h.Logged = append(h.Logged, r)
	}
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

// readPageDir decodes the page directory. Every directory entry sits
// right after the section it describes, so the Xilinx DBG entry is at
// offset+length, and the arena records follow that entry: one
// arenaRecLen byte record per arena in use. The header's arena table
// names the records, and the records are laid out in the order the
// arenas were first written, which is not always arena order:
// t7_gen_for has arena 2 first. So the only check is that every named
// record is at pageDirStart plus a multiple of arenaRecLen and that no
// two slots share one.
//
// A record names at most arenaMax pages. When an arena needs more, the
// record's first word points at a continuation record of the same
// layout, written among the pages right after the first page it names,
// and the chain goes on from there: t9_tr70000 has 117 pages, 100 in
// the record and 17 in one continuation.
//
// The result has one Arena per slot of the arena table. A slot that is
// zero, an arena no object was written to, gives an Arena with a zero
// Offset and no pages.
func readPageDir(d []byte, h *Header, dbg DirEntry) ([]Arena, error) {
	base := dbg.Offset + dbg.Length + pageDirStart
	arenas := make([]Arena, len(h.ArenaOffsets))
	seen := map[uint64]int{}
	for i, off := range h.ArenaOffsets {
		if off == 0 {
			continue
		}
		if off < base || (off-base)%arenaRecLen != 0 {
			return nil, fmt.Errorf("arena record %d at %#x is not %#x plus a multiple of %#x", i, off, base, arenaRecLen)
		}
		if j, dup := seen[off]; dup {
			return nil, fmt.Errorf("arena table slots %d and %d both name the record at %#x", j, i, off)
		}
		seen[off] = i
		a := Arena{Offset: off}
		for rec, k := off, 0; rec != 0; k++ {
			next, err := readArenaRecord(d, rec, i, k, &a)
			if err != nil {
				return nil, err
			}
			if next != 0 && next <= rec {
				return nil, fmt.Errorf("arena record %d continuation %d at %#x points back to %#x", i, k, rec, next)
			}
			rec = next
		}
		arenas[i] = a
	}
	// The records are contiguous: as many as there are used slots.
	for k := 0; k < len(seen); k++ {
		if _, ok := seen[base+uint64(k)*arenaRecLen]; !ok {
			return nil, fmt.Errorf("no arena table slot names the record at %#x", base+uint64(k)*arenaRecLen)
		}
	}
	return arenas, nil
}

// readArenaRecord appends the pages one arena record names to a, and
// returns the offset of the continuation record its first word names,
// 0 when there is none. A record lists its pages as uint64 offsets from
// arenaOffsets and uint32 compressed lengths from arenaLens, with the
// page count at arenaCount; the slots past the count are zero. i is
// the arena and k the position in the chain, for messages.
func readArenaRecord(d []byte, off uint64, i, k int, a *Arena) (uint64, error) {
	if off+arenaRecLen > uint64(len(d)) {
		return 0, fmt.Errorf("arena record %d continuation %d at %#x runs past the file", i, k, off)
	}
	r := d[off : off+arenaRecLen]
	next := binary.LittleEndian.Uint64(r[0:])
	n := binary.LittleEndian.Uint64(r[arenaCount:])
	if n > arenaMax {
		return 0, fmt.Errorf("arena record %d continuation %d names %d pages, more than %d", i, k, n, arenaMax)
	}
	if next != 0 && n != arenaMax {
		return 0, fmt.Errorf("arena record %d continuation %d names %d pages and still continues at %#x", i, k, n, next)
	}
	for j := uint64(0); j < n; j++ {
		ref := PageRef{
			Offset:        binary.LittleEndian.Uint64(r[arenaOffsets+8*j:]),
			CompressedLen: uint64(binary.LittleEndian.Uint32(r[arenaLens+4*j:])),
		}
		if ref.Offset+ref.CompressedLen > uint64(len(d)) {
			return 0, fmt.Errorf("arena %d page %d at %#x+%d runs past the file", i, len(a.Pages), ref.Offset, ref.CompressedLen)
		}
		a.Pages = append(a.Pages, ref)
	}
	for _, b := range r[arenaOffsets+8*n : arenaLens] {
		if b != 0 {
			return 0, fmt.Errorf("arena record %d continuation %d has a non-zero offset slot past page %d", i, k, n)
		}
	}
	for _, b := range r[arenaLens+4*n : arenaCount] {
		if b != 0 {
			return 0, fmt.Errorf("arena record %d continuation %d has a non-zero length slot past page %d", i, k, n)
		}
	}
	return next, nil
}
