// SPDX-License-Identifier: Apache-2.0

// Tier 80: where the bytes of a subprogram's static values lie.
package main

import (
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
    generic (
        k : integer := 7
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    %(types)s
    function f(c : std_ulogic) return std_ulogic is
        %(locals)s
        variable v : integer := 0;
    begin
        %(uses)s
        return c;
    end function;
begin
    p: process
        variable a : integer := 0;
    begin
        wait for 50 ns;
        a := k;
        s <= f('1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

func kase(name, brief, types, locals, uses, differs string, extraLocals []*c.Obj) {
	axis := "static values in the handle space. " + brief +
		", to see whether the bytes of a static composite value push the objects that come after the signals."
	body := c.Fill(tb, "brief", brief, "axis", axis, "types", types, "locals", locals, "uses", uses)
	vars := append([]*c.Obj{
		c.O("scope", "tb.p", "name", "a", "type", "integer", "kind", "variable", "port", ""),
		c.O("scope", "tb.f", "name", "c", "type", "std_ulogic", "kind", "local", "port", "in")},
		extraLocals...)
	vars = append(vars, c.O("scope", "tb.f", "name", "v", "type", "integer", "kind", "local", "port", ""))
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files:   []c.File{{Name: "tb.ent.vhdl", Body: body}},
		Signals: []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)},
		Trs:     []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")},
		Extra: c.O("variables", vars, "generics", []*c.Obj{
			c.O("instance", "", "scope", "tb", "name", "k", "type", "integer", "value", "7")}),
		NoX: true})
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n",
			"    ],\n    xelab_args = [\n        \"-debug\",\n        \"typical\",\n        \"-debug\",\n        \"subprogram\",\n    ],\n)\n", 1)
	})
}

// local is one of the function's own declarations in the truth.
func local(name, typ string) *c.Obj {
	return c.O("scope", "tb.f", "name", name, "type", typ, "kind", "local", "port", "")
}

func main() {
	// The design with no static value: a scalar local only.
	kase("t80_stv_none____", "a function with a scalar local and no static value",
		"", "", "v := v + 1;", "t56_typ_arr_unus", nil)
	// A four element array local with a literal initial value, which
	// tier 56 measured at 0x10 of handle space.
	kase("t80_stv_arr4____", "an array local of four integers, initialised",
		"type arr_t is array (0 to 3) of integer;",
		"variable ar : arr_t := (others => 0);", "v := v + ar(1);", "t80_stv_none____",
		[]*c.Obj{local("ar", "arr_t")})
	// The same array with eight elements, at 0x20.
	kase("t80_stv_arr8____", "an array local of eight integers, initialised",
		"type arr_t is array (0 to 7) of integer;",
		"variable ar : arr_t := (others => 0);", "v := v + ar(1);", "t80_stv_arr4____",
		[]*c.Obj{local("ar", "arr_t")})
	// A record local of eight bytes, at 0xc.
	kase("t80_stv_rec8____", "a record local of eight bytes, initialised",
		`type rec_t is record
        m : integer;
        n : integer;
    end record;`,
		"variable rc : rec_t := (0, 0);", "v := v + rc.m;", "t80_stv_arr4____",
		[]*c.Obj{local("rc", "rec_t")})
	// Two array locals, at 0x20, to see whether two static values lie
	// together.
	kase("t80_stv_two_arr_", "two array locals of four integers",
		"type arr_t is array (0 to 3) of integer;",
		`variable ar : arr_t := (others => 0);
        variable br : arr_t := (others => 1);`, "v := v + ar(1) + br(2);", "t80_stv_arr4____",
		[]*c.Obj{local("ar", "arr_t"), local("br", "arr_t")})
}
