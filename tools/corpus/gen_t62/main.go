// SPDX-License-Identifier: Apache-2.0

// Tier 62: net strengths, pull sources, switches and gate primitives.
package main

import (
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
	t60 "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gent60"
)

// part is the declarations and the records a case adds to the design.
type part struct {
	sigs []*c.Obj
	trs  []*c.Obj
}

func (p part) plus(q part) part {
	return part{append(append([]*c.Obj{}, p.sigs...), q.sigs...),
		append(append([]*c.Obj{}, p.trs...), q.trs...)}
}

func kase(name, brief, decl, differs string, p part) {
	axis := "strength. " + brief + " beside a logic, under typical, to see whether a net's drive strength, a pull source, a switch or a gate primitive leaves anything in the declaration, the hierarchy or the records."
	body := c.Fill(t60.SV, "brief", brief, "axis", axis, "decl", decl, "write", "")
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig("tb", "s", "logic", 1)}, p.sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, p.trs...),
		NoX:     true})
}

// net declares a net that holds v0 from time 0 and v50 from 50 ns, after
// the all X record every net starts with. records pins the count of
// records the drivers write, where it is more than the changes.
func net(name, v0, v50, typ string, width, records int) part {
	sg := c.Sig("tb", name, typ, width)
	if records != 0 {
		sg.Set("records", records)
	}
	trs := []*c.Obj{c.Tr(0, name, strings.Repeat("X", width)), c.Tr(0, name, v0)}
	if v50 != v0 {
		trs = append(trs, c.Tr(50, name, v50))
	}
	return part{[]*c.Obj{sg}, trs}
}

// wire is a scalar wire, the usual case of net.
func wire(name, v0, v50 string, records int) part {
	return net(name, v0, v50, "wire", 1, records)
}

func main() {
	kase("t62_str_none____", "nothing", "", "t11_sv_logic____", part{})
	kase("t62_str_wire____", "a wire driven by s", "wire w; assign w = s;", "t62_str_none____", wire("w", "0", "1", 3))
	kase("t62_str_tri_____", "a tri driven by s", "tri w; assign w = s;", "t62_str_wire____", net("w", "0", "1", "tri", 1, 0))
	kase("t62_str_uwire___", "a uwire driven by s", "uwire w; assign w = s;", "t62_str_wire____", net("w", "0", "1", "uwire", 1, 0))
	kase("t62_str_pullup__", "a wire with a pullup", "wire w; pullup (w);", "t62_str_wire____", wire("w", "1", "1", 2))
	kase("t62_str_pulldn__", "a wire with a pulldown", "wire w; pulldown (w);", "t62_str_pullup__", wire("w", "0", "0", 2))
	kase("t62_str_pu_drv__", "a pullup under a driver that releases", "wire w; pullup (w); assign w = s ? 1'bz : 1'b0;", "t62_str_pullup__", wire("w", "0", "1", 4))
	kase("t62_str_weak____", "a weak 1 under a driver that releases", "wire w; assign (weak0, weak1) w = 1'b1; assign w = s ? 1'bz : 1'b0;", "t62_str_pu_drv__", wire("w", "0", "1", 4))
	kase("t62_str_strong__", "a strong driver over a weak one", "wire w; assign (weak0, weak1) w = 1'b0; assign (strong0, strong1) w = s;", "t62_str_wire____", wire("w", "0", "1", 4))
	kase("t62_str_equal___", "two strong drivers that disagree", "wire w; assign w = 1'b0; assign w = s;", "t62_str_strong__", wire("w", "0", "X", 3))
	kase("t62_str_mixed___", "a strong0 weak1 driver against a pull1", "wire w; assign (strong0, weak1) w = s; assign (pull0, pull1) w = 1'b1;", "t62_str_strong__", wire("w", "0", "1", 4))
	kase("t62_str_supply__", "a supply driver against a strong one", "wire w; assign (supply0, supply1) w = 1'b0; assign w = s;", "t62_str_strong__", wire("w", "0", "0", 4))
	kase("t62_str_wand____", "a wand with a weak 0 and a strong s", "wand w; assign (weak0, weak1) w = 1'b0; assign w = s;", "t62_str_strong__", net("w", "0", "1", "wand", 1, 4))
	kase("t62_str_bufif___", "a bufif1 gate", "wire w; bufif1 (w, 1'b1, s);", "t62_str_wire____", wire("w", "Z", "1", 4))
	kase("t62_str_bufif_n_", "a named bufif1 gate", "wire w; bufif1 g1 (w, 1'b1, s);", "t62_str_bufif___", wire("w", "Z", "1", 4))
	kase("t62_str_and_____", "an and gate", "wire w; and (w, s, 1'b1);", "t62_str_bufif___", wire("w", "0", "1", 3))
	kase("t62_str_and_2___", "two and gates in one statement", "wire w, x; and g1 (w, s, 1'b1), g2 (x, s, s);", "t62_str_and_____",
		wire("w", "0", "1", 0).plus(wire("x", "0", "1", 0)))
	kase("t62_str_nmos____", "an nmos switch", "wire w; nmos (w, 1'b1, s);", "t62_str_bufif___", wire("w", "Z", "1", 4))
	kase("t62_str_vec_pu__", "a vector with pullups", "wire [3:0] v; pullup p [3:0] (v); assign v = s ? 4'bzz01 : 4'b0000;", "t62_str_pu_drv__",
		part{[]*c.Obj{c.Sig("tb", "v", "wire", 4, "records", 13)},
			[]*c.Obj{c.Tr(0, "v", "XXXX"), c.Tr(0, "v", "XXX0"), c.Tr(0, "v", "XX00"), c.Tr(0, "v", "X000"), c.Tr(0, "v", "0000"),
				c.Tr(50, "v", "0001"), c.Tr(50, "v", "0101"), c.Tr(50, "v", "1101")}})
	kase("t62_str_vec_1drv", "a vector with one driver", "wire [3:0] v; assign v = s ? 4'b1101 : 4'b0000;", "t62_str_wire____", net("v", "0000", "1101", "wire", 4, 3))
	kase("t62_str_vec_2drv", "a vector with two drivers", "wire [3:0] v; assign v = s ? 4'bzz01 : 4'b0000; assign v = 4'bz1zz;", "t62_str_vec_1drv",
		part{[]*c.Obj{c.Sig("tb", "v", "wire", 4, "records", 9)},
			[]*c.Obj{c.Tr(0, "v", "XXXX"), c.Tr(0, "v", "XXX0"), c.Tr(0, "v", "XX00"), c.Tr(0, "v", "0X00"),
				c.Tr(50, "v", "0X01"), c.Tr(50, "v", "0101"), c.Tr(50, "v", "Z101")}})
	kase("t62_str_gate_dly", "a delayed and gate", "wire w; and #3 (w, s, 1'b1);", "t62_str_and_____",
		part{[]*c.Obj{c.Sig("tb", "w", "wire", 1)}, []*c.Obj{c.Tr(0, "w", "X"), c.Tr(3, "w", "0"), c.Tr(53, "w", "1")}})
}
