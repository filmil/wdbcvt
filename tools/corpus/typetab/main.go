// SPDX-License-Identifier: Apache-2.0

// Splits a type table into entries and prints each as hex, for
// exploration of the entries the reader does not parse yet.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

const magic = "Xilinx ISim TYPE FILE 001"

type entry struct {
	off  int
	body []byte
}

// entries finds the type table and splits it into its length prefixed
// entries. It returns the offset of the table, the count and size words
// of its header, the entries, and where they end.
func entries(b []byte) (int, uint32, uint32, []entry, int, error) {
	i := bytes.Index(b, []byte(magic))
	if i < 0 {
		return 0, 0, 0, nil, 0, fmt.Errorf("no type table")
	}
	n := binary.LittleEndian.Uint32(b[i+0x20:])
	size := binary.LittleEndian.Uint32(b[i+0x24:])
	p := i + 0x28
	var out []entry
	for k := uint32(0); k < n; k++ {
		ln := int(binary.LittleEndian.Uint32(b[p:]))
		out = append(out, entry{off: p - i, body: b[p+4 : p+ln]})
		p += ln
	}
	return i, n, size, out, p - i, nil
}

func show(fn string) error {
	b, err := os.ReadFile(fn)
	if err != nil {
		return err
	}
	_, n, size, es, end, err := entries(b)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	parts := strings.Split(fn, "/")
	name := fn
	if len(parts) > 1 {
		name = parts[len(parts)-2]
	}
	fmt.Printf("%s: %d entries, size word 0x%x, entries end at 0x%x\n", name, n, size, end)
	for k, e := range es {
		tag := e.body[0]
		z := bytes.IndexByte(e.body[4:], 0) + 4
		nm := string(e.body[4:z])
		rest := e.body[z+1:]
		words := "RAW " + hex.EncodeToString(rest)
		if len(rest)%4 == 0 {
			var w []string
			for i := 0; i+4 <= len(rest); i += 4 {
				w = append(w, fmt.Sprintf("%08x", binary.LittleEndian.Uint32(rest[i:])))
			}
			words = strings.Join(w, " ")
		}
		fmt.Printf("  [%d] kind 0x%02x %-12s %s\n", k, tag, "'"+nm+"'", words)
	}
	return nil
}

func main() {
	for _, fn := range os.Args[1:] {
		if err := show(fn); err != nil {
			fmt.Fprintln(os.Stderr, "typetab:", err)
			os.Exit(1)
		}
	}
}
