// SPDX-License-Identifier: Apache-2.0

// Tier 68: where the value of a SystemVerilog string goes.
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
        %(write)s
        #50 $finish;
    end
endmodule
`

const tcl = `open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
%s
close_vcd
exit
`

// opt holds what a case adds to the plain design: the objects it
// declares beside the logic and their records, the declarations the
// source has and the file does not, an xsim script, and whether it
// elaborates under -debug all.
type opt struct {
	sigs   []*c.Obj
	trs    []*c.Obj
	absent []*c.Obj
	dbg    bool
	tcl    []string
}

func kase(name, brief, decl, write, differs string, o opt) {
	axis := "string storage. " + brief + " beside a logic, to see whether the characters of a SystemVerilog string reach the database, and where."
	body := c.Fill(sv, "brief", brief, "axis", axis, "decl", decl, "write", write)
	var extra *c.Obj
	if o.absent != nil {
		extra = c.O("absent", o.absent)
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.sv", Body: body}},
		Signals: append([]*c.Obj{c.Sig("tb", "s", "logic", 1)}, o.sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, o.trs...),
		Extra:   extra, NoX: true})
	attrs := ""
	if o.tcl != nil {
		c.WriteFile(name, "xsim.tcl", strings.Replace(tcl, "%s", strings.Join(o.tcl, "\n"), 1))
		attrs += "    tcl = \"xsim.tcl\",\n"
	}
	if o.dbg {
		attrs += "    xelab_args = [\n        \"-debug\",\n        \"all\",\n    ],\n"
	}
	if attrs != "" {
		c.PatchBuild(name, func(b string) string {
			return strings.Replace(b, "    ],\n)\n", "    ],\n"+attrs+")\n", 1)
		})
	}
}

// absent names a declaration the source has and the file does not.
func absent(name, typ string) []*c.Obj {
	return []*c.Obj{c.O("scope", "tb", "name", name, "type", typ)}
}

// held is the declaration and the record of a string under -debug all,
// which tier 60 measured: one 32 bit record of zeros at time 0 and
// nothing at the write.
func held(names ...string) opt {
	var o opt
	for _, n := range names {
		o.sigs = append(o.sigs, c.Sig("tb", n, "string", 32))
		o.trs = append(o.trs, c.Tr(0, n, strings.Repeat("0", 32)))
	}
	return o
}

// The characters the cases carry. They are chosen to appear nowhere
// else in a database, so that a search of the file and of the inflated
// pages for them answers the question of the tier:
//
//	bazel run //tools/pagegrep -- -pat ZQXJ <the database>
const (
	lit4 = "ZQXJ"
	lit2 = "WPMK"
	// Forty characters, past the four bytes a string record holds and
	// past the eight bytes of a record.
	lit40 = "ZQXJWPMKZQXJWPMKZQXJWPMKZQXJWPMKZQXJWPMK"
)

func main() {
	kase("t68_str_lit4____", "a string variable of four characters",
		`string str = "`+lit4+`";`, `str = "`+lit2+`";`, "t11_sv_logic____",
		opt{absent: absent("str", "string")})
	kase("t68_str_lit40___", "a string variable of forty characters",
		`string str = "`+lit40+`";`, `str = "`+lit2+`";`, "t68_str_lit4____",
		opt{absent: absent("str", "string")})
	kase("t68_str_noinit__", "a string variable without an initializer",
		`string str;`, `str = "`+lit2+`";`, "t68_str_lit4____",
		opt{absent: absent("str", "string")})
	kase("t68_str_arr_____", "an unpacked array of two strings",
		`string a [0:1] = '{"`+lit4+`", "`+lit2+`"};`, `a[1] = "`+lit4+`";`, "t68_str_lit4____",
		opt{absent: absent("a", "string")})
	kase("t68_str_log_____", "the string named in log_wave under typical",
		`string str = "`+lit4+`";`, `str = "`+lit2+`";`, "t68_str_lit4____",
		opt{absent: absent("str", "string"), tcl: []string{"log_wave /tb/str", "run -all"}})
	kase("t68_str_dbg_____", "the four character string under -debug all",
		`string str = "`+lit4+`";`, `str = "`+lit2+`";`, "t68_str_lit4____",
		func() opt { o := held("str"); o.dbg = true; return o }())
	kase("t68_str_dbg40___", "the forty character string under -debug all",
		`string str = "`+lit40+`";`, `str = "`+lit2+`";`, "t68_str_dbg_____",
		func() opt { o := held("str"); o.dbg = true; return o }())
	// The array is one object of two placeholders, not two objects.
	kase("t68_str_dbg_arr_", "the array of two strings under -debug all",
		`string a [0:1] = '{"`+lit4+`", "`+lit2+`"};`, `a[1] = "`+lit4+`";`, "t68_str_dbg_____",
		opt{dbg: true,
			sigs: []*c.Obj{c.Sig("tb", "a", "memory", 64, "elements", 2, "element_width", 32, "element_type", "string")},
			trs:  []*c.Obj{c.Tr(0, "a", "("+strings.Repeat("0", 32)+", "+strings.Repeat("0", 32)+")")}})
	// The control of the tier: an unpacked array of bytes holding the
	// same characters does put them in a record, so a search that finds
	// nothing in the string cases is a fact about strings and not about
	// the search.
	kase("t68_str_byte____", "an unpacked array of four bytes holding the same characters",
		`byte b [0:3] = '{"Z", "Q", "X", "J"};`, `b[0] = "W";`, "t68_str_lit4____",
		opt{sigs: []*c.Obj{c.Sig("tb", "b", "memory", 32, "elements", 4, "element_width", 8, "element_type", "byte")},
			trs: []*c.Obj{
				c.Tr(0, "b", "(90, 81, 88, 74)"),
				c.Tr(50, "b", "(87, 81, 88, 74)")}})
}
