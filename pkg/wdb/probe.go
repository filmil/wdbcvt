// SPDX-License-Identifier: Apache-2.0

// Package wdb holds what is known about the Vivado xsim waveform database
// (.wdb) container format, and the tools used to find out more.
//
// The format is undocumented by AMD. Nothing here guesses: Probe reports
// measurements taken from a file, and docs/wdb-format.md records which of
// those measurements have hardened into facts.
package wdb

import (
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"sort"
	"unicode"
)

// HeaderLen is how many leading bytes Probe keeps verbatim in a Report.
const HeaderLen = 64

// blockLen is the window Probe measures entropy over.
const blockLen = 4096

// minStringLen is the shortest run of printable bytes Probe reports.
const minStringLen = 6

// Block is one fixed-size window of the file with its Shannon entropy, in
// bits per byte. Values near 8 mean compressed or encrypted content;
// values near 4 or below mean structured or textual content.
type Block struct {
	Offset  int64
	Length  int
	Entropy float64
}

// StringRun is a run of printable ASCII found in the file.
type StringRun struct {
	Offset int64
	Text   string
}

// Report is everything Probe measured about one file.
type Report struct {
	// Size is the file size in bytes.
	Size int64
	// Header is the first HeaderLen bytes, or the whole file when it is
	// shorter.
	Header []byte
	// Blocks holds the per-window entropy, in file order.
	Blocks []Block
	// Strings holds the printable runs, in file order.
	Strings []StringRun
	// Histogram counts every byte value over the whole file.
	Histogram [256]int64
}

// Probe measures the file behind r without interpreting it. It reads the
// whole file once. size must be the file's size in bytes.
func Probe(r io.Reader, size int64) (*Report, error) {
	if size < 0 {
		return nil, fmt.Errorf("negative size: %d", size)
	}
	rep := &Report{Size: size}

	buf := make([]byte, blockLen)
	var offset int64
	var run []byte
	var runStart int64

	flushRun := func() {
		if len(run) >= minStringLen {
			rep.Strings = append(rep.Strings, StringRun{
				Offset: runStart,
				Text:   string(run),
			})
		}
		run = run[:0]
	}

	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			chunk := buf[:n]

			if int64(len(rep.Header)) < HeaderLen {
				want := HeaderLen - len(rep.Header)
				if want > n {
					want = n
				}
				rep.Header = append(rep.Header, chunk[:want]...)
			}

			for i, b := range chunk {
				rep.Histogram[b]++
				if isPrintable(b) {
					if len(run) == 0 {
						runStart = offset + int64(i)
					}
					run = append(run, b)
				} else {
					flushRun()
				}
			}

			rep.Blocks = append(rep.Blocks, Block{
				Offset:  offset,
				Length:  n,
				Entropy: entropy(chunk),
			})
			offset += int64(n)
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading at offset %d: %w", offset, err)
		}
	}
	flushRun()

	return rep, nil
}

// MeanEntropy is the entropy of the whole file, in bits per byte.
func (r *Report) MeanEntropy() float64 {
	var total int64
	for _, c := range r.Histogram {
		total += c
	}
	if total == 0 {
		return 0
	}
	var h float64
	for _, c := range r.Histogram {
		if c == 0 {
			continue
		}
		p := float64(c) / float64(total)
		h -= p * math.Log2(p)
	}
	return h
}

// TopStrings returns at most n of the longest printable runs, longest
// first. Ties keep file order.
func (r *Report) TopStrings(n int) []StringRun {
	out := make([]StringRun, len(r.Strings))
	copy(out, r.Strings)
	sort.SliceStable(out, func(i, j int) bool {
		return len(out[i].Text) > len(out[j].Text)
	})
	if n < len(out) {
		out = out[:n]
	}
	return out
}

// HeaderDump renders the kept header bytes as a hex dump.
func (r *Report) HeaderDump() string {
	return hex.Dump(r.Header)
}

func isPrintable(b byte) bool {
	return b < unicode.MaxASCII && unicode.IsPrint(rune(b))
}

func entropy(b []byte) float64 {
	if len(b) == 0 {
		return 0
	}
	var counts [256]int
	for _, c := range b {
		counts[c]++
	}
	var h float64
	for _, c := range counts {
		if c == 0 {
			continue
		}
		p := float64(c) / float64(len(b))
		h -= p * math.Log2(p)
	}
	return h
}
