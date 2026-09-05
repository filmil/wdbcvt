// SPDX-License-Identifier: Apache-2.0

// Tier 72: what each -debug mode writes, and what the flag bytes mean.
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
    int r;

    function automatic int f(input int x);
        int tmp;
        tmp = x + 1;
        return tmp;
    endfunction

    initial begin
        #50 s = 1'b1;
        r = f(1);
        #50 $finish;
    end
endmodule
`

type opt struct {
	sigs   []*c.Obj
	trs    []*c.Obj
	absent []*c.Obj
}

func kase(name, brief, differs string, xelab []string, o opt) {
	axis := "debug modes. The same design, a logic beside a function with a local, elaborated with " + brief +
		", to see what the mode writes and which of the flag bytes of header words 14 and 15 it sets."
	body := c.Fill(sv, "brief", brief, "axis", axis)
	extra := c.O()
	if o.absent != nil {
		extra.Set("absent", o.absent)
	}
	if len(extra.Keys()) == 0 {
		extra = nil
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig("tb", "s", "logic", 1)}, o.sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, o.trs...),
		Extra:   extra, NoX: true})
	attrs := "    xelab_args = [\n"
	for _, a := range xelab {
		attrs += "        \"" + a + "\",\n"
	}
	attrs += "    ],\n"
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n", "    ],\n"+attrs+")\n", 1)
	})
}

// result is the int the function's result goes into, which every mode
// records.
func result() opt {
	return opt{sigs: []*c.Obj{c.Sig("tb", "r", "int", 32)},
		trs: []*c.Obj{c.Tr(0, "r", "0"), c.Tr(50, "r", "2")}}
}

// withLocals is a mode that writes the subprogram's own declarations:
// the return variable, the input and the local, none of them logged.
// A SystemVerilog subprogram declares them as objects of the function
// scope, where a VHDL one lists them as variables, t22_dbg_subprog.
func withLocals() opt {
	o := result()
	v := func(name, port string) *c.Obj {
		sg := c.Sig("tb.f", name, "int", 32, "logged", false)
		if port != "" {
			sg.Set("port", port)
		}
		return sg
	}
	o.sigs = append(o.sigs, v("f", ""), v("x", "in"), v("tmp", ""))
	return o
}

func main() {
	kase("t72_dbg_typical_", "-debug typical", "t11_sv_logic____", []string{"-debug", "typical"}, result())
	kase("t72_dbg_wave____", "-debug wave", "t72_dbg_typical_", []string{"-debug", "wave"}, result())
	// A narrow mode on its own writes no database at all: the run
	// stops with `ERROR: [Simulator 45-10] The current simulation was
	// compiled without trace information`, so each is paired with
	// `wave`, as the tier 24 cases are.
	kase("t72_dbg_line____", "-debug wave -debug line", "t72_dbg_wave____", []string{"-debug", "wave", "-debug", "line"}, withLocals())
	kase("t72_dbg_subprog_", "-debug wave -debug subprogram", "t72_dbg_line____", []string{"-debug", "wave", "-debug", "subprogram"}, withLocals())
	kase("t72_dbg_all_____", "-debug all", "t72_dbg_typical_", []string{"-debug", "all"}, withLocals())
	kase("t72_dbg_drivers_", "-debug wave -debug drivers", "t72_dbg_wave____", []string{"-debug", "wave", "-debug", "drivers"}, result())
	kase("t72_dbg_readers_", "-debug wave -debug readers", "t72_dbg_drivers_", []string{"-debug", "wave", "-debug", "readers"}, result())
}
