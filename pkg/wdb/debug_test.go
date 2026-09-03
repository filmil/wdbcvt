// SPDX-License-Identifier: Apache-2.0

package wdb

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHeaderCountsAreChecked raises header word 0, the range record
// count, of a corpus database by one and expects the reader to reject
// the file: the count no longer fits region 4. The unmodified file is
// read first, so the test fails loudly if the corpus moves.
func TestHeaderCountsAreChecked(t *testing.T) {
	path := filepath.Join(corpusCases(t)[0], "sim.wdb")
	d, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Read(d); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	i := bytes.Index(d, []byte(debugMagic))
	if i < 0 {
		t.Fatalf("%s: no %q", path, debugMagic)
	}
	// The header words follow the magic, the timestamp and precision
	// words, the 18 offsets and the 4 counts.
	at := i + len(debugMagic) + 1 + 8 + 4*dbgOffsets + 4*dbgCounts
	w := binary.LittleEndian.Uint32(d[at:])
	binary.LittleEndian.PutUint32(d[at:], w+1)
	_, err = Read(d)
	if err == nil || !strings.Contains(err.Error(), "DBG header word 0") {
		t.Fatalf("header word 0 raised from %d to %d: got %v, want an error naming the word", w, w+1, err)
	}
}
