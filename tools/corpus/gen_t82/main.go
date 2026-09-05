// SPDX-License-Identifier: Apache-2.0

// Tier 82: what a package body adds to the handle space.
package main

import (
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

const pkg = `
--! @file
--! @brief Corpus case: %(brief)s
--!
--! Axis: %(axis)s

package pk is
%(decls)s
end package;

package body pk is
%(body)s
    function f return integer is
    begin
        return %(ret)s;
    end function;
end package body;
`

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
begin
    p: process
        variable a : integer := 0;
    begin
        wait for 50 ns;
        a := k + work.pk.f;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

func kase(name, brief, decls, body, ret, differs string, vars []*c.Obj) {
	axis := "a package body in the handle space. " + brief +
		", to see whether what a package body declares lands in the package's block or past the second region."
	files := []c.File{
		{Name: "pk.pkg.vhdl", Body: c.Fill(pkg, "brief", brief, "axis", axis,
			"decls", decls, "body", body, "ret", ret)},
		{Name: "tb.ent.vhdl", Body: c.Fill(tb, "brief", brief, "axis", axis)},
	}
	all := append([]*c.Obj{
		c.O("scope", "tb.p", "name", "a", "type", "integer", "kind", "variable", "port", "")}, vars...)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs, Files: files,
		Signals: []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)},
		Trs:     []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")},
		Extra: c.O("variables", all, "generics", []*c.Obj{
			c.O("instance", "", "scope", "tb", "name", "k", "type", "integer", "value", "7")}),
		NoX: true})
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    srcs = [\n",
			"    # Compilation order matters to xvhdl; do not sort.\n    # do not sort\n    srcs = [\n", 1)
	})
}

// con is a package constant of the truth: an object with no record.
func con(name, typ, value string) *c.Obj {
	return c.O("scope", "pk", "name", name, "type", typ, "kind", "constant",
		"value", value, "logged", false)
}

func main() {
	// A package with a function and an empty body, the baseline.
	kase("t82_pkb_none____", "a package whose body declares nothing",
		"    function f return integer;", "", "3", "t81_pkt_fn______", nil)
	// One constant in the body rather than in the package.
	kase("t82_pkb_1con____", "a package whose body declares one integer constant",
		"    function f return integer;", "    constant b0 : integer := 1;", "b0",
		"t82_pkb_none____", []*c.Obj{con("b0", "integer", "1")})
	// Four of them.
	kase("t82_pkb_4con____", "a package whose body declares four integer constants",
		"    function f return integer;",
		`    constant b0 : integer := 1;
    constant b1 : integer := 2;
    constant b2 : integer := 3;
    constant b3 : integer := 4;`, "b0 + b1 + b2 + b3", "t82_pkb_1con____",
		[]*c.Obj{con("b0", "integer", "1"), con("b1", "integer", "2"),
			con("b2", "integer", "3"), con("b3", "integer", "4")})
	// A table in the body, which is the shape a library package has.
	kase("t82_pkb_arr16___", "a package whose body declares a constant array of sixteen integers",
		"    function f return integer;",
		`    type arr_t is array (0 to 15) of integer;
    constant t : arr_t := (others => 7);`, "t(0)", "t82_pkb_1con____",
		[]*c.Obj{con("t", "arr_t", "(7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7)")})
	// A static composite value inside the package's own function,
	// which is what a library package's bodies are full of.
	kase("t82_pkb_fn_stat_", "a package whose function has an array local with a literal value",
		"    function f return integer;",
		`    type arr_t is array (0 to 3) of integer;

    function g return integer is
        variable ar : arr_t := (others => 2);
    begin
        return ar(1);
    end function;`, "g", "t82_pkb_none____", nil)
	// A deferred constant: named in the package, valued in the body.
	kase("t82_pkb_deferred", "a package with a deferred constant",
		`    constant d : integer;
    function f return integer;`, "    constant d : integer := 5;", "d",
		"t82_pkb_none____", []*c.Obj{con("d", "integer", "5")})
}
