// SPDX-License-Identifier: Apache-2.0

// Tier 77: what a package costs the handle space, item by item.
package main

import (
	"fmt"
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
%(body)s`

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
begin
    p: process
        variable a : integer := 0;
    begin
        a := %(read)s;
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

// con is a package constant of the truth, which is an object with no
// record.
func con(name, typ, value string) *c.Obj {
	return c.O("scope", "pk", "name", name, "type", typ, "kind", "constant",
		"value", value, "logged", false)
}

func kase(name, brief, decls, body, read, differs string, vars []*c.Obj) {
	axis := "package handle space. A package that declares " + brief +
		", read from the process, to see what each kind of declaration costs the handle space."
	all := append([]*c.Obj{
		c.O("scope", "tb.p", "name", "a", "type", "integer", "kind", "variable")}, vars...)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{
			{Name: "pk.pkg.vhdl", Body: c.Fill(pkg, "brief", brief, "axis", axis, "decls", decls, "body", body)},
			{Name: "tb.ent.vhdl", Body: c.Fill(tb, "brief", brief, "axis", axis, "read", read)},
		},
		Signals: []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)},
		Trs:     []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")},
		Extra:   c.O("variables", all), NoX: true})
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    srcs = [\n",
			"    # Compilation order matters to xvhdl; do not sort.\n    # do not sort\n    srcs = [\n", 1)
	})
}

// ints declares n integer constants and names them in the truth.
func ints(n int) (string, []*c.Obj) {
	var d strings.Builder
	var vars []*c.Obj
	for i := 0; i < n; i++ {
		fmt.Fprintf(&d, "    constant c%d : integer := %d;\n", i, i+1)
		vars = append(vars, con(fmt.Sprintf("c%d", i), "integer", fmt.Sprint(i+1)))
	}
	return strings.TrimRight(d.String(), "\n"), vars
}

func main() {
	// One integer constant, the shape tier 54 measured, as this tier's
	// own baseline.
	d, v := ints(1)
	kase("t77_pkc_1con____", "one integer constant", d, "", "work.pk.c0", "t54_pkg_con_var_", v)
	d, v = ints(4)
	kase("t77_pkc_4con____", "four integer constants", d, "", "work.pk.c0", "t77_pkc_1con____", v)
	d, v = ints(16)
	kase("t77_pkc_16con___", "sixteen integer constants", d, "", "work.pk.c0", "t77_pkc_4con____", v)

	// A real constant, eight bytes where an integer is four.
	kase("t77_pkc_1real___", "one real constant",
		"    constant r : real := 1.5;", "", "integer(work.pk.r)", "t77_pkc_1con____",
		[]*c.Obj{con("r", "real", "1.5")})

	// A constant array of sixteen integers, one object of 64 bytes.
	kase("t77_pkc_arr16___", "a constant array of sixteen integers",
		`    type arr_t is array (0 to 15) of integer;
    constant t : arr_t := (others => 7);`, "", "work.pk.t(0)", "t77_pkc_16con___",
		[]*c.Obj{con("t", "arr_t", "(7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7)")})

	// Subprograms, which declare no object of their own.
	kase("t77_pkc_1fn_____", "one function and no constant",
		`    function f return integer;`,
		`
package body pk is
    function f return integer is
    begin
        return 3;
    end function;
end package body;
`, "work.pk.f", "t77_pkc_1con____", nil)
	kase("t77_pkc_4fn_____", "four functions and no constant",
		`    function f return integer;
    function g return integer;
    function h return integer;
    function i return integer;`,
		`
package body pk is
    function f return integer is
    begin
        return 3;
    end function;
    function g return integer is
    begin
        return 4;
    end function;
    function h return integer is
    begin
        return 5;
    end function;
    function i return integer is
    begin
        return 6;
    end function;
end package body;
`, "work.pk.f", "t77_pkc_1fn_____", nil)

	// A type and nothing else, which is read through a qualified value.
	kase("t77_pkc_1type___", "one type and no object",
		"    type small_t is range 0 to 7;", "", "integer(work.pk.small_t'(3))", "t77_pkc_1con____", nil)
}
