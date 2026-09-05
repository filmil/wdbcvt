// SPDX-License-Identifier: Apache-2.0

// Tier 76: the storage class of the subprogram objects no case has
// declared, in search of the class 5 that has never been seen.
package main

import (
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const vhdl = `
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
    %(decl)s
begin
    p: process
        %(local)s
    begin
        wait for 50 ns;
        %(body)s
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

const sv = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int r = 0;
    %(decl)s

    initial begin
        #50 s = 1'b1;
        %(body)s
        #50 $finish;
    end
endmodule
`

type opt struct {
	sigs      []*c.Obj
	trs       []*c.Obj
	variables []*c.Obj
	absent    []*c.Obj
	sv        bool
	decl      string
	local     string
	body      string
}

func kase(name, brief, differs string, o opt) {
	axis := "storage classes. " + brief +
		", to see which storage class word 28 of the instance record gives it, and whether any form gives the 5 that no case has produced."
	var files []c.File
	sigs := []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)}
	trs := []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}
	if o.sv {
		files = []c.File{{Name: "tb.sv", Body: c.Fill(sv, "brief", brief, "axis", axis, "decl", o.decl, "body", o.body)}}
		sigs = []*c.Obj{c.Sig("tb", "s", "logic", 1), c.Sig("tb", "r", "int", 32)}
		trs = []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1"), c.Tr(0, "r", "0")}
		// Every case leaves r at 2 at 50 ns; a case that passes
		// through another value first says so in trs.
	} else {
		files = []c.File{{Name: "tb.ent.vhdl",
			Body: c.Fill(vhdl, "brief", brief, "axis", axis, "decl", o.decl, "local", o.local, "body", o.body)}}
	}
	extra := c.O()
	if o.variables != nil {
		extra.Set("variables", o.variables)
	}
	if o.absent != nil {
		extra.Set("absent", o.absent)
	}
	if len(extra.Keys()) == 0 {
		extra = nil
	}
	all := append(trs, o.trs...)
	if o.sv {
		all = append(all, c.Tr(50, "r", "2"))
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs, Files: files,
		Signals: append(sigs, o.sigs...), Trs: all, Extra: extra, NoX: true})
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n",
			"    ],\n    xelab_args = [\n        \"-debug\",\n        \"typical\",\n        \"-debug\",\n        \"subprogram\",\n    ],\n)\n", 1)
	})
}

func main() {
	// A file as a parameter and as a local of a procedure, which no
	// case has declared inside a subprogram.
	kase("t76_stc_file_prm", "a file parameter of a procedure", "t51_sub_file_prm", opt{
		decl: `type int_file is file of integer;
    procedure w(file f : int_file; v : integer) is
    begin
        write(f, v);
    end procedure;`,
		local: "file lf : int_file open write_mode is \"t76a.txt\";",
		body:  "w(lf, 5);",
		variables: []*c.Obj{
			c.O("scope", "tb.w", "name", "v", "type", "integer", "kind", "local", "port", "in"),
			c.O("scope", "tb.p", "name", "lf", "type", "int_file", "kind", "variable", "port", "")},
		// The file parameter itself leaves no object at all.
		absent: []*c.Obj{c.O("scope", "tb.w", "name", "f", "type", "int_file")}})
	kase("t76_stc_file_loc", "a file local of a procedure", "t76_stc_file_prm", opt{
		decl: `type int_file is file of integer;
    procedure w(v : integer) is
        file f : int_file open write_mode is "t76b.txt";
    begin
        write(f, v);
    end procedure;`,
		body: "w(5);",
		variables: []*c.Obj{
			c.O("scope", "tb.w", "name", "v", "type", "integer", "kind", "local", "port", "in")},
		absent: []*c.Obj{c.O("scope", "tb.w", "name", "f", "type", "int_file")}})
	// An access parameter, where the local of the same type is class 3.
	kase("t76_stc_acc_prm_", "an access parameter of a procedure", "t50_sub_acc_loc_", opt{
		decl: `type int_ptr is access integer;
    procedure w(p : inout int_ptr) is
    begin
        p := new integer'(5);
    end procedure;`,
		local: "variable q : int_ptr;",
		body:  "w(q);\n        deallocate(q);",
		variables: []*c.Obj{
			c.O("scope", "tb.w", "name", "p", "type", "int_ptr", "kind", "local", "port", "inout"),
			c.O("scope", "tb.p", "name", "q", "type", "int_ptr", "kind", "variable", "port", "")}})
	// SystemVerilog arguments and locals of the shapes the corpus has
	// not put in a subprogram.
	kase("t76_stc_sv_ref__", "a ref argument of a function", "t72_dbg_subprog_", opt{sv: true,
		decl: `function automatic void bump(ref int x);
        x = x + 1;
    endfunction`,
		body: "bump(r); bump(r);",
		// Two calls, so r passes through 1 on its way to 2.
		trs:  []*c.Obj{c.Tr(50, "r", "1")},
		sigs: []*c.Obj{c.Sig("tb.bump", "x", "int", 32, "logged", false)}})
	kase("t76_stc_sv_out__", "an output argument of a task", "t76_stc_sv_ref__", opt{sv: true,
		decl: `task automatic two(output int y);
        y = 2;
    endtask`,
		body: "two(r);",
		sigs: []*c.Obj{c.Sig("tb.two", "y", "int", 32, "port", "out", "logged", false)}})
	kase("t76_stc_sv_arr__", "an unpacked array local of a function", "t76_stc_sv_ref__", opt{sv: true,
		decl: `function automatic int sum2();
        int a[2];
        a[0] = 1;
        a[1] = 1;
        return a[0] + a[1];
    endfunction`,
		body: "r = sum2();",
		sigs: []*c.Obj{
			c.Sig("tb.sum2", "sum2", "int", 32, "logged", false),
			c.Sig("tb.sum2", "a", "", 64, "logged", false)}})
	kase("t76_stc_sv_str__", "a string local of a function", "t76_stc_sv_arr__", opt{sv: true,
		decl: `function automatic int len();
        string t;
        t = "ab";
        return t.len();
    endfunction`,
		body: "r = len();",
		sigs: []*c.Obj{c.Sig("tb.len", "len", "int", 32, "logged", false)},
		absent: []*c.Obj{c.O("scope", "tb.len", "name", "t", "type", "string")}})
}
