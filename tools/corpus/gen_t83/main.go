// SPDX-License-Identifier: Apache-2.0

// Tier 83: the four bytes a record adds over its declared size.
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
    begin
        wait for 50 ns;
        s <= f('1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

func kase(name, brief, types, locals, uses, differs string, extraLocals []*c.Obj) {
	axis := "the four bytes of a record. " + brief +
		", to see how much handle space the static value of a record adds beyond the bytes the record declares."
	body := c.Fill(tb, "brief", brief, "axis", axis, "types", types, "locals", locals, "uses", uses)
	vars := append([]*c.Obj{
		c.O("scope", "tb.f", "name", "c", "type", "std_ulogic", "kind", "local", "port", "in")},
		extraLocals...)
	vars = append(vars, c.O("scope", "tb.f", "name", "v", "type", "integer", "kind", "local", "port", ""))
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

func local(name, typ string) *c.Obj {
	return c.O("scope", "tb.f", "name", name, "type", typ, "kind", "local", "port", "")
}

func main() {
	// One integer in the record: eight bytes declared.
	kase("t83_rec_1int____", "a record of one integer",
		`type rec_t is record
        n : integer;
    end record;`, "variable r : rec_t := (n => 0);", "v := v + r.n;",
		"t56_typ_rec_loc_", []*c.Obj{local("r", "rec_t")})
	// One std_ulogic, the smallest field there is.
	kase("t83_rec_1sul____", "a record of one std_ulogic",
		`type rec_t is record
        a : std_ulogic;
    end record;`, "variable r : rec_t := (a => '0');", "if r.a = '1' then v := v + 1; end if;",
		"t83_rec_1int____", []*c.Obj{local("r", "rec_t")})
	// Four and eight integers, to see whether the four bytes scale.
	kase("t83_rec_4int____", "a record of four integers",
		`type rec_t is record
        m : integer;
        n : integer;
        o : integer;
        p : integer;
    end record;`, "variable r : rec_t := (0, 0, 0, 0);", "v := v + r.m;",
		"t83_rec_1int____", []*c.Obj{local("r", "rec_t")})
	kase("t83_rec_8int____", "a record of eight integers",
		`type rec_t is record
        a1 : integer;
        a2 : integer;
        a3 : integer;
        a4 : integer;
        a5 : integer;
        a6 : integer;
        a7 : integer;
        a8 : integer;
    end record;`, "variable r : rec_t := (0, 0, 0, 0, 0, 0, 0, 0);", "v := v + r.a1;",
		"t83_rec_4int____", []*c.Obj{local("r", "rec_t")})
	// A record inside a record, to see whether each one adds four.
	kase("t83_rec_nested__", "a record holding another record",
		`type inner_t is record
        n : integer;
    end record;
    type rec_t is record
        i : inner_t;
        m : integer;
    end record;`, "variable r : rec_t := (i => (n => 0), m => 0);", "v := v + r.m;",
		"t83_rec_1int____", []*c.Obj{local("r", "rec_t")})
	// An array inside a record, and an array of records.
	kase("t83_rec_arr4____", "a record holding an array of four integers",
		`type arr_t is array (0 to 3) of integer;
    type rec_t is record
        t : arr_t;
    end record;`, "variable r : rec_t := (t => (others => 0));", "v := v + r.t(1);",
		"t83_rec_1int____", []*c.Obj{local("r", "rec_t")})
	// Four records in an array, and two records inside one, to tell
	// four bytes once per value from four bytes per record.
	kase("t83_rec_arr4rec_", "an array of four records",
		`type rec_t is record
        n : integer;
    end record;
    type arr_t is array (0 to 3) of rec_t;`, "variable r : arr_t := (others => (n => 0));",
		"v := v + r(1).n;", "t83_rec_arr_of__", []*c.Obj{local("r", "arr_t")})
	kase("t83_rec_2nested_", "a record holding two records",
		`type inner_t is record
        n : integer;
    end record;
    type rec_t is record
        i : inner_t;
        j : inner_t;
    end record;`, "variable r : rec_t := (i => (n => 0), j => (n => 0));", "v := v + r.i.n;",
		"t83_rec_nested__", []*c.Obj{local("r", "rec_t")})
	kase("t83_rec_arr_of__", "an array of two records",
		`type rec_t is record
        n : integer;
    end record;
    type arr_t is array (0 to 1) of rec_t;`, "variable r : arr_t := (others => (n => 0));",
		"v := v + r(1).n;", "t83_rec_1int____", []*c.Obj{local("r", "arr_t")})
}
