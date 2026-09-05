// SPDX-License-Identifier: Apache-2.0

// Tier 64: several partial drivers on one net.
package main

import (
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
module child(input i, output o);
    assign o = i;
endmodule
`

const child4 = `
module child4(input a, input b, output c, output d);
    assign c = a;
    assign d = b;
endmodule
`

const child2 = child + `
module child2(input i, output o);
    assign o = ~i;
endmodule
`

const bidi = `
module bidi(inout io);
    assign io = 1'b0;
endmodule
`

func kase(name, brief, decl, differs string, sigs, trs []*c.Obj, kid string) {
	axis := "several partial drivers. " + brief + " beside a logic, under typical, to see the order and the place of the records the drivers write."
	body := c.Fill(sv, "brief", brief, "axis", axis, "decl", decl, "child", kid)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig("tb", "s", "logic", 1)}, sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, trs...),
		NoX:     true})
}

// net declares the driven net, with the record count the truth pins.
func net(w, records int, name string) *c.Obj {
	return c.Sig("tb", name, "wire", w, "records", records)
}

// pt is one record of a sequence, and seq turns a list of them into the
// records of one signal.
type pt struct {
	t int
	v string
}

func seq(name string, ps ...pt) []*c.Obj {
	out := make([]*c.Obj, len(ps))
	for i, p := range ps {
		out[i] = c.Tr(p.t, name, p.v)
	}
	return out
}

// kidSigs and kidTrs are the ports of one instance of child, whose input
// takes iv at time 0 and ov at 50 ns, and whose output follows it.
func kidSigs(u string, records int) []*c.Obj {
	return []*c.Obj{
		c.Sig("tb."+u, "i", "wire", 1, "port", "in", "records", 3),
		c.Sig("tb."+u, "o", "wire", 1, "port", "out", "records", records),
	}
}

func kidTrs(u, iv, ov string) []*c.Obj {
	return append(seq("tb."+u+".i", pt{0, "X"}, pt{0, iv}, pt{50, ov}),
		seq("tb."+u+".o", pt{0, "X"}, pt{0, iv}, pt{50, ov})...)
}

// kid4Sigs and kid4Trs are the four ports of one instance of child4.
func kid4Sigs(u string) []*c.Obj {
	return []*c.Obj{
		c.Sig("tb."+u, "a", "wire", 1, "port", "in", "records", 3),
		c.Sig("tb."+u, "b", "wire", 1, "port", "in", "records", 3),
		c.Sig("tb."+u, "c", "wire", 1, "port", "out", "records", 4),
		c.Sig("tb."+u, "d", "wire", 1, "port", "out", "records", 4),
	}
}

func kid4Trs(u string) []*c.Obj {
	var out []*c.Obj
	for _, n := range []string{"a", "b", "c", "d"} {
		out = append(out, seq("tb."+u+"."+n, pt{0, "X"}, pt{0, "0"}, pt{50, "1"})...)
	}
	return out
}

// rise is the record of a scalar net that holds 0 and goes to 1.
func rise(name string) []*c.Obj {
	return seq(name, pt{0, "X"}, pt{0, "0"}, pt{50, "1"})
}

func cat(ls ...[]*c.Obj) []*c.Obj {
	var out []*c.Obj
	for _, l := range ls {
		out = append(out, l...)
	}
	return out
}

func main() {
	z62 := strings.Repeat("Z", 62)
	z2398 := strings.Repeat("Z", 2398)

	kase("t64_ord_src_rev_", "the two drivers of two bits in the other source order", "wire [3:0] v; assign v[3] = ~s; assign v[0] = s;", "t63_pdr_two_bits",
		[]*c.Obj{net(4, 5, "v")}, seq("v", pt{0, "XZZX"}, pt{0, "1ZZX"}, pt{0, "1ZZ0"}, pt{50, "0ZZ0"}, pt{50, "0ZZ1"}), "")
	kase("t64_ord_gen4____", "four drivers of four bits from a generate loop", "wire [3:0] v; genvar i; for (i = 0; i < 4; i = i + 1) begin : g assign v[i] = s; end", "t63_pdr_two_bits",
		[]*c.Obj{net(4, 9, "v")}, seq("v", pt{0, "XXXX"}, pt{0, "XXX0"}, pt{0, "XX00"}, pt{0, "X000"}, pt{0, "0000"}, pt{50, "0001"}, pt{50, "0011"}, pt{50, "0111"}, pt{50, "1111"}), "")
	kase("t64_ord_gen_rev_", "the generate loop counting down", "wire [3:0] v; genvar i; for (i = 3; i >= 0; i = i - 1) begin : g assign v[i] = s; end", "t64_ord_gen4____",
		[]*c.Obj{net(4, 9, "v")}, seq("v", pt{0, "XXXX"}, pt{0, "0XXX"}, pt{0, "00XX"}, pt{0, "000X"}, pt{0, "0000"}, pt{50, "1000"}, pt{50, "1100"}, pt{50, "1110"}, pt{50, "1111"}), "")
	kase("t64_ord_w64_two_", "two drivers in the two pairs of a 64 bit net", "wire [63:0] v; assign v[0] = s; assign v[63] = ~s;", "t63_pdr_two_bits",
		[]*c.Obj{net(64, 5, "v")}, seq("v", pt{0, "X" + z62 + "X"}, pt{0, "X" + z62 + "0"}, pt{0, "1" + z62 + "0"}, pt{50, "1" + z62 + "1"}, pt{50, "0" + z62 + "1"}), "")
	kase("t64_ord_2400_two", "two drivers in the first and last chunk of a 2400 bit net", "wire [2399:0] v; assign v[0] = s; assign v[2399] = ~s;", "t64_ord_w64_two_",
		[]*c.Obj{net(2400, 5, "v")}, seq("v", pt{0, "X" + z2398 + "X"}, pt{0, "X" + z2398 + "0"}, pt{0, "1" + z2398 + "0"}, pt{50, "1" + z2398 + "1"}, pt{50, "0" + z2398 + "1"}), "")
	kase("t64_ord_unp_elem", "a driver of one bit of one element of an unpacked array of nets", "wire [3:0] v [0:1]; assign v[1][2] = s;", "t63_pdr_bit0____",
		[]*c.Obj{c.Sig("tb", "v", "memory", 8, "elements", 2, "element_width", 4, "element_type", "wire", "records", 3)},
		seq("v", pt{0, "(ZZZZ, ZXZZ)"}, pt{0, "(ZZZZ, Z0ZZ)"}, pt{50, "(ZZZZ, Z1ZZ)"}), "")
	kase("t64_ord_unp_whol", "a driver of one whole element of an unpacked array of nets", "wire [3:0] v [0:1]; assign v[1] = {4{s}};", "t64_ord_unp_elem",
		[]*c.Obj{c.Sig("tb", "v", "memory", 8, "elements", 2, "element_width", 4, "element_type", "wire", "records", 3)},
		seq("v", pt{0, "(ZZZZ, XXXX)"}, pt{0, "(ZZZZ, 0000)"}, pt{50, "(ZZZZ, 1111)"}), "")
	kase("t64_ord_two_kids", "two child outputs on two bits of one net", "wire [3:0] v; child u0(.i(s), .o(v[1])); child u1(.i(~s), .o(v[3]));", "t63_pdr_port_bit",
		cat([]*c.Obj{net(4, 7, "v")}, kidSigs("u0", 7), kidSigs("u1", 7)),
		cat(seq("v", pt{0, "XZXZ"}, pt{0, "XZ0Z"}, pt{0, "1Z0Z"}, pt{50, "1Z1Z"}, pt{50, "0Z1Z"}),
			kidTrs("u0", "0", "1"), kidTrs("u1", "1", "0")), child)
	kase("t64_ord_pos_expr", "one child output on a scalar net, the input bound to an expression", "wire w; child u1(.i(~s), .o(w));", "t64_ord_two_kids",
		cat([]*c.Obj{net(1, 4, "w")}, kidSigs("u1", 4)),
		cat(seq("w", pt{0, "X"}, pt{0, "1"}, pt{50, "0"}), kidTrs("u1", "1", "0")), child)
	kase("t64_ord_pos_bit3", "one child output on bit 3, the input bound to the logic", "wire [3:0] v; child u1(.i(s), .o(v[3]));", "t64_ord_two_kids",
		cat([]*c.Obj{net(4, 4, "v")}, kidSigs("u1", 4)),
		cat(seq("v", pt{0, "XZZZ"}, pt{0, "0ZZZ"}, pt{50, "1ZZZ"}), kidTrs("u1", "0", "1")), child)
	kase("t64_ord_two_same", "two child outputs on two bits, both inputs bound to the logic", "wire [3:0] v; child u0(.i(s), .o(v[1])); child u1(.i(s), .o(v[3]));", "t64_ord_two_kids",
		cat([]*c.Obj{net(4, 7, "v")}, kidSigs("u0", 7), kidSigs("u1", 7)),
		cat(seq("v", pt{0, "XZXZ"}, pt{0, "XZ0Z"}, pt{0, "0Z0Z"}, pt{50, "1Z0Z"}, pt{50, "1Z1Z"}),
			kidTrs("u0", "0", "1"), kidTrs("u1", "0", "1")), child)
	kase("t64_ord_two_nets", "two child outputs on two scalar nets", "wire a, b; child u0(.i(s), .o(a)); child u1(.i(~s), .o(b));", "t64_ord_two_kids",
		cat([]*c.Obj{net(1, 4, "a"), net(1, 4, "b")}, kidSigs("u0", 4), kidSigs("u1", 4)),
		cat(rise("a"), seq("b", pt{0, "X"}, pt{0, "1"}, pt{50, "0"}), kidTrs("u0", "0", "1"), kidTrs("u1", "1", "0")), child)
	kase("t64_ord_two_pos4", "two instances of a child with four ports", "wire c0, d0, c1, d1; child4 u0(.a(s), .b(s), .c(c0), .d(d0)); child4 u1(.a(s), .b(s), .c(c1), .d(d1));", "t64_ord_two_nets",
		cat([]*c.Obj{net(1, 4, "c0"), net(1, 4, "d0"), net(1, 4, "c1"), net(1, 4, "d1")}, kid4Sigs("u0"), kid4Sigs("u1")),
		cat(rise("c0"), rise("d0"), rise("c1"), rise("d1"), kid4Trs("u0"), kid4Trs("u1")), child4)
	kase("t64_ord_three___", "three instances of the child on three scalar nets", "wire a, b, c; child u0(.i(s), .o(a)); child u1(.i(s), .o(b)); child u2(.i(s), .o(c));", "t64_ord_two_nets",
		cat([]*c.Obj{net(1, 4, "a"), net(1, 4, "b"), net(1, 4, "c")}, kidSigs("u0", 4), kidSigs("u1", 4), kidSigs("u2", 4)),
		cat(rise("a"), rise("b"), rise("c"), kidTrs("u0", "0", "1"), kidTrs("u1", "0", "1"), kidTrs("u2", "0", "1")), child)
	kase("t64_ord_two_mods", "two children of two modules on two scalar nets", "wire a, b; child u0(.i(s), .o(a)); child2 u1(.i(s), .o(b));", "t64_ord_two_nets",
		cat([]*c.Obj{net(1, 4, "a"), net(1, 4, "b")}, kidSigs("u0", 4), kidSigs("u1", 4)),
		cat(rise("a"), seq("b", pt{0, "X"}, pt{0, "1"}, pt{50, "0"}), kidTrs("u0", "0", "1"),
			seq("tb.u1.i", pt{0, "X"}, pt{0, "0"}, pt{50, "1"}), seq("tb.u1.o", pt{0, "X"}, pt{0, "1"}, pt{50, "0"})), child2)
	kase("t64_ord_gen_kids", "two child outputs on two bits from a generate loop", "wire [3:0] v; genvar i; for (i = 0; i < 2; i = i + 1) begin : g child u(.i(s), .o(v[i * 3])); end", "t64_ord_two_same",
		cat([]*c.Obj{net(4, 7, "v")}, kidSigs("g[0].u", 7), kidSigs("g[1].u", 7)),
		cat(seq("v", pt{0, "XZZX"}, pt{0, "XZZ0"}, pt{0, "0ZZ0"}, pt{50, "1ZZ0"}, pt{50, "1ZZ1"}),
			kidTrs("g[0].u", "0", "1"), kidTrs("g[1].u", "0", "1")), child)
	kase("t64_ord_inout___", "a child inout on one bit beside a driver of another", "wire [3:0] v; assign v[1] = s; bidi u(.io(v[3]));", "t63_pdr_port_bit",
		[]*c.Obj{net(4, 5, "v"), c.Sig("tb.u", "io", "wire", 1, "port", "inout", "records", 5)},
		cat(seq("v", pt{0, "XZXZ"}, pt{0, "XZ0Z"}, pt{0, "0Z0Z"}, pt{50, "0Z1Z"}),
			seq("tb.u.io", pt{0, "X"}, pt{0, "0"})), bidi)
	kase("t64_ord_self____", "a driver of one bit from another bit of the same net", "wire [3:0] v; assign v[0] = s; assign v[1] = v[0];", "t63_pdr_two_bits",
		[]*c.Obj{net(4, 6, "v")}, seq("v", pt{0, "ZZXX"}, pt{0, "ZZX0"}, pt{0, "ZZ00"}, pt{50, "ZZ01"}, pt{50, "ZZ11"}), "")
	kase("t64_ord_chain___", "a driver of one bit of a second net from a bit of the first", "wire [3:0] v, w; assign v[0] = s; assign w[1] = v[0];", "t64_ord_self____",
		[]*c.Obj{net(4, 3, "v"), net(4, 3, "w")},
		cat(seq("v", pt{0, "ZZZX"}, pt{0, "ZZZ0"}, pt{50, "ZZZ1"}), seq("w", pt{0, "ZZXZ"}, pt{0, "ZZ0Z"}, pt{50, "ZZ1Z"})), "")
}
