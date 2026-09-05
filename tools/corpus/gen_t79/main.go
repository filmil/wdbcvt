// SPDX-License-Identifier: Apache-2.0

// Tier 79: what lies below the first signal's handle.
package main

import (
	"fmt"
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const tb = `
--! @file
--! @brief Corpus case: %(brief)s
--!
--! Axis: %(axis)s

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
%(subs)s
begin
    p: process
    begin
        wait for 50 ns;
%(calls)s
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

// proc writes a procedure with n integer variables, each assigned so
// that the compiler keeps it.
func proc(name string, n int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "    procedure %s(v : integer) is\n", name)
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "        variable a%d : integer := 0;\n", i)
	}
	b.WriteString("    begin\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "        a%d := v + %d;\n", i, i)
	}
	if n == 0 {
		b.WriteString("        null;\n")
	}
	b.WriteString("    end procedure;")
	return b.String()
}

// locals names the variables of a procedure in the truth.
func locals(scope string, n int) []*c.Obj {
	var out []*c.Obj
	for i := 0; i < n; i++ {
		out = append(out, c.O("scope", scope, "name", fmt.Sprintf("a%d", i),
			"type", "integer", "kind", "local", "port", ""))
	}
	return out
}

func kase(name, brief, subs, calls, differs string, vars []*c.Obj) {
	axis := "the space below the first signal. " + brief +
		", to see what lies under the handle `0x768` that every first signal has, and whether filling it moves the signal."
	body := c.Fill(tb, "brief", brief, "axis", axis, "subs", subs, "calls", calls)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.ent.vhdl", Body: body}},
		Signals: []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)},
		Trs:     []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")},
		Extra:   c.O("variables", vars), NoX: true})
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n",
			"    ],\n    xelab_args = [\n        \"-debug\",\n        \"typical\",\n        \"-debug\",\n        \"subprogram\",\n    ],\n)\n", 1)
	})
}

func main() {
	// One procedure, with more and more variables in its frame.
	for _, n := range []int{1, 16, 64, 256, 512} {
		vars := append([]*c.Obj{
			c.O("scope", "tb.w", "name", "v", "type", "integer", "kind", "local", "port", "in")},
			locals("tb.w", n)...)
		kase(fmt.Sprintf("t79_frm_%03dint__", n), fmt.Sprintf("a procedure with %d integer variables", n),
			proc("w", n), "        w(5);", "t50_sub_in_var__", vars)
	}
	// Two procedures, to see whether their frames share the space.
	vars := append(append([]*c.Obj{
		c.O("scope", "tb.w", "name", "v", "type", "integer", "kind", "local", "port", "in")},
		locals("tb.w", 64)...),
		append([]*c.Obj{
			c.O("scope", "tb.x", "name", "v", "type", "integer", "kind", "local", "port", "in")},
			locals("tb.x", 64)...)...)
	kase("t79_frm_two_64__", "two procedures with 64 integer variables each",
		proc("w", 64)+"\n\n"+proc("x", 64), "        w(5);\n        x(6);", "t79_frm_064int__", vars)
}
