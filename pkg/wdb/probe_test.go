// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

func TestProbeHeaderAndSize(t *testing.T) {
	data := bytes.Repeat([]byte{0xAB}, 100)
	rep, err := Probe(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if rep.Size != 100 {
		t.Errorf("Size = %d, want 100", rep.Size)
	}
	if len(rep.Header) != HeaderLen {
		t.Errorf("len(Header) = %d, want %d", len(rep.Header), HeaderLen)
	}
}

func TestProbeShortFileKeepsWholeHeader(t *testing.T) {
	data := []byte("short")
	rep, err := Probe(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if !bytes.Equal(rep.Header, data) {
		t.Errorf("Header = %q, want %q", rep.Header, data)
	}
}

func TestProbeFindsStrings(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0, 1, 2})
	buf.WriteString("counter_types")
	buf.Write([]byte{0, 0})
	buf.WriteString("tb")
	buf.Write([]byte{0})

	rep, err := Probe(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if len(rep.Strings) != 1 {
		t.Fatalf("Strings = %v, want exactly the one long run", rep.Strings)
	}
	if got := rep.Strings[0]; got.Text != "counter_types" || got.Offset != 3 {
		t.Errorf("Strings[0] = %+v, want offset 3 text counter_types", got)
	}
}

func TestProbeSpansBlockBoundary(t *testing.T) {
	// A string that straddles the block window must still be found whole.
	pad := bytes.Repeat([]byte{0}, blockLen-4)
	data := append(append([]byte{}, pad...), []byte("straddling_name")...)

	rep, err := Probe(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	found := false
	for _, s := range rep.Strings {
		if s.Text == "straddling_name" {
			found = true
		}
	}
	if !found {
		t.Errorf("Strings = %v, want it to contain straddling_name", rep.Strings)
	}
}

func TestMeanEntropy(t *testing.T) {
	zeros := bytes.Repeat([]byte{0}, 8192)
	rep, err := Probe(bytes.NewReader(zeros), int64(len(zeros)))
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if h := rep.MeanEntropy(); h != 0 {
		t.Errorf("MeanEntropy of a constant file = %v, want 0", h)
	}

	// Every byte value equally often: exactly 8 bits per byte.
	var flat []byte
	for i := 0; i < 32; i++ {
		for v := 0; v < 256; v++ {
			flat = append(flat, byte(v))
		}
	}
	rep, err = Probe(bytes.NewReader(flat), int64(len(flat)))
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if h := rep.MeanEntropy(); math.Abs(h-8) > 1e-9 {
		t.Errorf("MeanEntropy of a uniform file = %v, want 8", h)
	}
}

func TestTopStrings(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("shorter")
	buf.WriteByte(0)
	buf.WriteString("a_much_longer_identifier")
	buf.WriteByte(0)

	rep, err := Probe(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	top := rep.TopStrings(1)
	if len(top) != 1 || top[0].Text != "a_much_longer_identifier" {
		t.Errorf("TopStrings(1) = %v, want the longest run first", top)
	}
}

func TestHeaderDump(t *testing.T) {
	rep, err := Probe(bytes.NewReader([]byte("abc")), 3)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if !strings.Contains(rep.HeaderDump(), "abc") {
		t.Errorf("HeaderDump() = %q, want it to show the bytes", rep.HeaderDump())
	}
}
