// SPDX-License-Identifier: Apache-2.0

// Tier 58: what log_wave can name, in SystemVerilog.
package main

import (
	"fmt"
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const tb = `// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps

module tb;
    typedef struct packed { logic a; logic [3:0] b; } st_t;
    parameter P = 3;
    localparam L = 4;
    logic s = 1'b0;
    logic [3:0] v = 4'b0000;
    logic [3:0] m [0:1] = '{4'd0, 4'd0};
    st_t st = '{a: 1'b0, b: 4'b0000};
    int i = 7;
    real r = 0.5;
    genvar g;
    for (g = 0; g < 2; g++) begin : gb
        wire gw = s;
    end
    task inc(input int x);
        int tmp;
        tmp = x + 1;
        i = tmp;
    endtask
    initial begin : blk
        int bv;
        bv = 1;
        #10;
        s = 1'b1;
        v = 4'b0101;
        m[1] = 4'd9;
        st = '{a: 1'b1, b: 4'b0011};
        r = 1.5;
        inc(1);
        bv = 2;
        #10;
        $finish;
    end
endmodule
`

const tcl = `open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
%s
run -all
close_vcd
exit
`

// logged names the objects the truth expects records for. Everything
// else is declared with "logged": false.
type logged []string

func (l logged) has(p string) bool {
	for _, e := range l {
		if e == p {
			return true
		}
	}
	return false
}

// lg is the key an object carries when the script does not log it.
func (l logged) lg(p string) []any {
	if l.has(p) {
		return nil
	}
	return []any{"logged", false}
}

func kase(name, brief, axis string, lines []string, lgd logged, differs string) {
	body := c.Fill(tb, "brief", brief, "axis", axis)
	signals := []*c.Obj{
		c.Sig("tb", "s", "logic", 1).With(lgd.lg("tb.s")...),
		c.Sig("tb", "v", "logic", 4).With(lgd.lg("tb.v")...),
		c.Sig("tb", "m", "memory", 8, "elements", 2, "element_width", 4, "element_type", "logic").With(lgd.lg("tb.m")...),
		c.O("scope", "tb", "name", "st", "type", "st_t", "width", 5, "fields", []*c.Obj{
			c.O("name", "a", "width", 1, "type", "logic"),
			c.O("name", "b", "width", 4, "type", "logic")}).With(lgd.lg("tb.st")...),
		c.Sig("tb", "i", "int", 32).With(lgd.lg("tb.i")...),
		c.Sig("tb", "r", "real", 64).With(lgd.lg("tb.r")...),
		c.Sig("tb.gb[0]", "gw", "wire", 1).With(lgd.lg("tb.gb[0].gw")...),
		c.Sig("tb.gb[1]", "gw", "wire", 1).With(lgd.lg("tb.gb[1].gw")...),
		c.Sig("tb.blk", "bv", "int", 32).With(lgd.lg("tb.blk.bv")...),
		c.Sig("tb.inc", "x", "int", 32, "port", "in").With(lgd.lg("tb.inc.x")...),
		c.Sig("tb.inc", "tmp", "int", 32).With(lgd.lg("tb.inc.tmp")...),
	}
	var trs []*c.Obj
	add := func(p string, xs ...*c.Obj) {
		if lgd.has(p) {
			trs = append(trs, xs...)
		}
	}
	add("tb.s", c.Tr(0, "s", "0"), c.Tr(10, "s", "1"))
	add("tb.v", c.Tr(0, "v", "0000"), c.Tr(10, "v", "0101"))
	add("tb.m", c.Tr(0, "m", "(0000, 0000)"), c.Tr(10, "m", "(0000, 1001)"))
	add("tb.st", c.Tr(0, "st.a", "0"), c.Tr(0, "st.b", "0000"), c.Tr(10, "st.a", "1"), c.Tr(10, "st.b", "0011"))
	add("tb.i", c.Tr(0, "i", "7"), c.Tr(10, "i", "2"))
	add("tb.r", c.Tr(0, "r", "0.5"), c.Tr(10, "r", "1.5"))
	for k := 0; k < 2; k++ {
		p := fmt.Sprintf("tb.gb[%d].gw", k)
		add(p, c.Tr(0, p, "X"), c.Tr(0, p, "0"), c.Tr(10, p, "1"))
	}
	add("tb.blk.bv", c.Tr(0, "tb.blk.bv", "0"), c.Tr(0, "tb.blk.bv", "1"), c.Tr(10, "tb.blk.bv", "2"))
	add("tb.inc.x", c.Tr(0, "tb.inc.x", "0"), c.Tr(10, "tb.inc.x", "1"))
	add("tb.inc.tmp", c.Tr(0, "tb.inc.tmp", "0"), c.Tr(10, "tb.inc.tmp", "2"))
	generics := []*c.Obj{
		c.O("instance", "", "scope", "tb", "name", "P", "type", "", "declared", "parameter",
			"value", c.Bits(3, 32)).With(lgd.lg("tb.P")...),
		c.O("instance", "", "scope", "tb", "name", "L", "type", "", "declared", "localparam",
			"value", c.Bits(4, 32)).With(lgd.lg("tb.L")...),
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{{Name: "tb.sv", Body: body}}, Signals: signals, Trs: trs, End: 20,
		Extra: c.O("generics", generics), NoX: true})
	c.WriteFile(name, "xsim.tcl", strings.Replace(tcl, "%s", strings.Join(lines, "\n"), 1))
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n", "    ],\n    tcl = \"xsim.tcl\",\n)\n", 1)
	})
}

var all = logged{"tb.s", "tb.v", "tb.m", "tb.st", "tb.i", "tb.r", "tb.gb[0].gw", "tb.gb[1].gw",
	"tb.blk.bv", "tb.inc.x", "tb.inc.tmp", "tb.P", "tb.L"}
var top = logged{"tb.s", "tb.v", "tb.m", "tb.st", "tb.i", "tb.r", "tb.gb[0].gw", "tb.gb[1].gw", "tb.P", "tb.L"}

const briefFmt = "log_wave naming %s of a SystemVerilog design with every kind of object"
const axisFmt = "logging. log_wave names %s, in a SystemVerilog design with a logic, a vector, a memory, a packed struct, an int, a real, a parameter, a localparam, a generate with a wire, a named block with a variable and a static task, to see what the database logs."

// one writes a case whose script logs a single object. vcd names what
// log_vcd takes when that is not the object itself, and the empty string
// means the case logs nothing to the VCD.
func one(name, what, obj string, lgd logged, vcd string, noVCD bool) {
	lines := []string{"log_wave " + obj}
	if !noVCD {
		v := vcd
		if v == "" {
			v = obj
		}
		lines = append([]string{"log_vcd " + v}, lines...)
	}
	kase(name, fmt.Sprintf(briefFmt, what), fmt.Sprintf(axisFmt, what), lines, lgd, "t58_sv_log_all__")
}

func main() {
	kase("t58_sv_log_all__", fmt.Sprintf(briefFmt, "everything, -recursive *"), fmt.Sprintf(axisFmt, "everything with -recursive *"),
		[]string{"log_vcd [get_objects -r /* ]", "log_wave -recursive *"}, all, "t57_log_all_____")
	kase("t58_sv_log_none_", fmt.Sprintf(briefFmt, "nothing"), fmt.Sprintf(axisFmt, "nothing, the script has no log_wave"),
		nil, nil, "t58_sv_log_all__")
	one("t58_sv_log_bit__", "one bit of a vector", "{/tb/v[3]}", nil, "", true)
	one("t58_sv_log_slc__", "a slice of a vector", "{/tb/v[2:1]}", logged{"tb.v"}, "", false)
	one("t58_sv_log_mem_e", "one element of a memory", "{/tb/m[1]}", nil, "", true)
	one("t58_sv_log_mem__", "a memory", "/tb/m", logged{"tb.m"}, "", false)
	one("t58_sv_log_st_fl", "a field of a packed struct", "/tb/st.a", nil, "", true)
	one("t58_sv_log_st___", "a packed struct", "/tb/st", logged{"tb.st"}, "", false)
	one("t58_sv_log_int__", "an int variable of the module", "/tb/i", logged{"tb.i"}, "", false)
	one("t58_sv_log_real_", "a real variable of the module", "/tb/r", logged{"tb.r"}, "", false)
	one("t58_sv_log_prm__", "a parameter", "/tb/P", logged{"tb.P"}, "", false)
	one("t58_sv_log_lprm_", "a localparam", "/tb/L", logged{"tb.L"}, "", false)
	one("t58_sv_log_blkv_", "a variable of a named block", "/tb/blk/bv", logged{"tb.blk.bv"}, "", false)
	one("t58_sv_log_blk__", "a named block", "/tb/blk", logged{"tb.blk.bv"}, "[get_objects /tb/blk/*]", false)
	one("t58_sv_log_tsk_l", "a local of a static task", "/tb/inc/tmp", logged{"tb.inc.tmp"}, "", false)
	one("t58_sv_log_tsk_a", "an argument of a static task", "/tb/inc/x", logged{"tb.inc.x"}, "", false)
	one("t58_sv_log_tsk__", "a static task", "/tb/inc", logged{"tb.inc.x", "tb.inc.tmp"}, "[get_objects /tb/inc/*]", false)
	one("t58_sv_log_gen_w", "the wire of one generate block, through get_objects -regexp",
		`[get_objects -regexp {/tb/.*gb\[1\].*}]`, logged{"tb.gb[1].gw"}, "", false)
	one("t58_sv_log_gen__", "a generate block by path", "{/tb/gb[1]}", nil, "", true)
	one("t58_sv_log_top__", "the top without -recursive", "/tb", top, "[get_objects /tb/*]", false)
}
