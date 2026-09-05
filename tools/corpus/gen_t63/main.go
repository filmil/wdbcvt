// SPDX-License-Identifier: Apache-2.0

// Tier 63: partial drivers on a net.
package main

import (
	"fmt"
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const sv = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    %(decl)s

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
%(child)s`

const child = `
module child(input i, output %so);
    assign o = %s;
endmodule
`

func childOf(w int) string {
	if w == 1 {
		return fmt.Sprintf(child, "", "i")
	}
	return fmt.Sprintf(child, fmt.Sprintf("[%d:0] ", w-1), fmt.Sprintf("{%d{i}}", w))
}

// val gives the value of a net whose bits in drv are driven by the
// expression named there, and Z everywhere else. s is 0, 1, or -1 for
// the first record, which holds X on the driven bits.
func val(w int, drv map[int]string, s int) string {
	out := make([]string, w)
	for i := range out {
		out[i] = "Z"
	}
	for b, e := range drv {
		switch {
		case s < 0:
			out[w-1-b] = "X"
		case e == "s":
			out[w-1-b] = fmt.Sprint(s)
		default:
			out[w-1-b] = fmt.Sprint(1 - s)
		}
	}
	return strings.Join(out, "")
}

// opt holds what a case adds beside the driven net: the records of the
// net when they are not the plain X, 0, 1 sequence, the record count the
// truth pins, another name for the net, a child module, and the
// declarations and records of that child's ports.
type opt struct {
	child     int
	extraSigs []*c.Obj
	extraTrs  []*c.Obj
	records   int
	vname     string
	trs       []*c.Obj
}

func kase(name, brief, decl, differs string, w int, drv map[int]string, o opt) {
	axis := "partial drivers. " + brief + " beside a logic, under typical, to see whether a driver of a bit, a slice or a port bound to part of a net records the whole net or the part."
	kid := ""
	if o.child != 0 {
		kid = childOf(o.child)
	}
	body := c.Fill(sv, "brief", brief, "axis", axis, "decl", decl, "child", kid)
	vname := o.vname
	if vname == "" {
		vname = "v"
	}
	sigs := []*c.Obj{c.Sig("tb", "s", "logic", 1)}
	if w != 0 {
		net := c.Sig("tb", vname, "wire", w)
		if o.records != 0 {
			net.Set("records", o.records)
		}
		sigs = append(sigs, net)
	}
	sigs = append(sigs, o.extraSigs...)
	t := []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}
	if w != 0 {
		t = append(t, c.Tr(0, vname, val(w, drv, -1)))
		if o.trs != nil {
			t = append(t, o.trs...)
		} else {
			t = append(t, c.Tr(0, vname, val(w, drv, 0)), c.Tr(50, vname, val(w, drv, 1)))
		}
	}
	t = append(t, o.extraTrs...)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{{Name: "tb.sv", Body: body}}, Signals: sigs, Trs: t, NoX: true})
}

// bit names one driven bit, and rng a driven range of them.
func bit(b int) map[int]string { return map[int]string{b: "s"} }

func rng(lo, hi int) map[int]string {
	d := map[int]string{}
	for b := lo; b <= hi; b++ {
		d[b] = "s"
	}
	return d
}

// port and ptrs give the declarations and the records of a child whose
// input is bound to the logic and whose output is w bits wide.
func port(w int) []*c.Obj {
	return []*c.Obj{
		c.Sig("tb.u", "i", "wire", 1, "port", "in", "records", 3),
		c.Sig("tb.u", "o", "wire", w, "port", "out", "records", 4),
	}
}

func ptrs(w int) []*c.Obj {
	return []*c.Obj{
		c.Tr(0, "tb.u.i", "X"), c.Tr(0, "tb.u.i", "0"), c.Tr(50, "tb.u.i", "1"),
		c.Tr(0, "tb.u.o", strings.Repeat("X", w)), c.Tr(0, "tb.u.o", strings.Repeat("0", w)),
		c.Tr(50, "tb.u.o", strings.Repeat("1", w)),
	}
}

func main() {
	kase("t63_pdr_bit0____", "bit 0 of a 4 bit net driven", "wire [3:0] v; assign v[0] = s;", "t62_str_vec_1drv", 4, bit(0), opt{records: 3})
	kase("t63_pdr_bit3____", "bit 3 of a 4 bit net driven", "wire [3:0] v; assign v[3] = s;", "t63_pdr_bit0____", 4, bit(3), opt{records: 3})
	kase("t63_pdr_two_bits", "bits 0 and 3 driven from two assigns", "wire [3:0] v; assign v[0] = s; assign v[3] = ~s;", "t63_pdr_bit0____", 4, map[int]string{0: "s", 3: "~s"},
		opt{records: 5, trs: []*c.Obj{c.Tr(0, "v", "XZZ0"), c.Tr(0, "v", "1ZZ0"), c.Tr(50, "v", "1ZZ1"), c.Tr(50, "v", "0ZZ1")}})
	kase("t63_pdr_slice___", "the low nibble of an 8 bit net driven", "wire [7:0] v; assign v[3:0] = {4{s}};", "t63_pdr_bit0____", 8, rng(0, 3), opt{records: 3})
	kase("t63_pdr_w64_bit0", "bit 0 of a 64 bit net driven", "wire [63:0] v; assign v[0] = s;", "t63_pdr_bit0____", 64, bit(0), opt{records: 3})
	kase("t63_pdr_w64_bit6", "bit 63 of a 64 bit net driven", "wire [63:0] v; assign v[63] = s;", "t63_pdr_w64_bit0", 64, bit(63), opt{records: 3})
	kase("t63_pdr_w64_hi__", "the high word of a 64 bit net driven", "wire [63:0] v; assign v[63:32] = {32{s}};", "t63_pdr_w64_bit0", 64, rng(32, 63), opt{records: 3})
	kase("t63_pdr_w64_all_", "all 64 bits of a 64 bit net driven", "wire [63:0] v; assign v = {64{s}};", "t63_pdr_w64_hi__", 64, rng(0, 63), opt{records: 3})
	kase("t63_pdr_2400_bit", "bit 0 of a 2400 bit net driven", "wire [2399:0] v; assign v[0] = s;", "t63_pdr_w64_bit0", 2400, bit(0), opt{records: 3})
	kase("t63_pdr_2400_hi_", "the top 400 bits of a 2400 bit net driven", "wire [2399:0] v; assign v[2399:2000] = {400{s}};", "t63_pdr_2400_bit", 2400, rng(2000, 2399), opt{records: 3})
	kase("t63_pdr_2400_all", "all 2400 bits of a 2400 bit net driven", "wire [2399:0] v; assign v = {2400{s}};", "t63_pdr_2400_hi_", 2400, rng(0, 2399), opt{records: 3})
	kase("t63_pdr_concat__", "two scalar nets driven through a concatenation", "wire a, b; assign {a, b} = {s, ~s};", "t62_str_wire____", 0, nil,
		opt{extraSigs: []*c.Obj{c.Sig("tb", "a", "wire", 1, "records", 3), c.Sig("tb", "b", "wire", 1, "records", 3)},
			extraTrs: []*c.Obj{c.Tr(0, "a", "X"), c.Tr(0, "a", "0"), c.Tr(50, "a", "1"),
				c.Tr(0, "b", "X"), c.Tr(0, "b", "1"), c.Tr(50, "b", "0")}})
	kase("t63_pdr_port_bit", "a child output bound to bit 1 of a 4 bit net", "wire [3:0] v; child u(.i(s), .o(v[1]));", "t63_pdr_bit0____", 4, bit(1),
		opt{child: 1, extraSigs: port(1), extraTrs: ptrs(1), records: 4})
	kase("t63_pdr_port_slc", "a child output bound to the high nibble of an 8 bit net", "wire [7:0] v; child u(.i(s), .o(v[7:4]));", "t63_pdr_port_bit", 8, rng(4, 7),
		opt{child: 4, extraSigs: port(4), extraTrs: ptrs(4), records: 4})
	kase("t63_pdr_port_hi_", "a child output bound to the high word of a 64 bit net", "wire [63:0] v; child u(.i(s), .o(v[63:32]));", "t63_pdr_port_slc", 64, rng(32, 63),
		opt{child: 32, extraSigs: port(32), extraTrs: ptrs(32), records: 4})
}
