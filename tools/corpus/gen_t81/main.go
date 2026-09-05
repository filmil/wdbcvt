// SPDX-License-Identifier: Apache-2.0

// Tier 81: where a package's handle space sits, before the objects that
// follow the signals or past them.
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

package %(name)s is
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
        a := k%(read)s;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

// con is a package constant of the truth: an object with no record.
func con(scope, name, typ, value string) *c.Obj {
	return c.O("scope", scope, "name", name, "type", typ, "kind", "constant",
		"value", value, "logged", false)
}

func kase(name, brief, read, differs string, pkgs []c.File, vars []*c.Obj) {
	axis := "where a package sits in the handle space. " + brief +
		", read from the process, to see whether the package moves the generic and the process variable that come after the signals."
	files := append([]c.File{}, pkgs...)
	files = append(files, c.File{Name: "tb.ent.vhdl",
		Body: c.Fill(tb, "brief", brief, "axis", axis, "read", read)})
	all := append([]*c.Obj{
		c.O("scope", "tb.p", "name", "a", "type", "integer", "kind", "variable", "port", "")}, vars...)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs, Files: files,
		Signals: []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)},
		Trs:     []*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")},
		Extra: c.O("variables", all, "generics", []*c.Obj{
			c.O("instance", "", "scope", "tb", "name", "k", "type", "integer", "value", "7")}),
		NoX: true})
	if len(pkgs) > 0 {
		c.PatchBuild(name, func(b string) string {
			return strings.Replace(b, "    srcs = [\n",
				"    # Compilation order matters to xvhdl; do not sort.\n    # do not sort\n    srcs = [\n", 1)
		})
	}
}

// pack writes one package file with n integer constants, or with a
// function when n is zero.
func pack(fileName, pkgName, brief, axis string, n int) c.File {
	var d strings.Builder
	body := ""
	if n == 0 {
		d.WriteString("    function f return integer;")
		body = fmt.Sprintf(`
package body %s is
    function f return integer is
    begin
        return 3;
    end function;
end package body;
`, pkgName)
	}
	for i := 0; i < n; i++ {
		fmt.Fprintf(&d, "    constant c%d : integer := %d;\n", i, i+1)
	}
	return c.File{Name: fileName, Body: c.Fill(pkg, "brief", brief, "axis", axis,
		"name", pkgName, "decls", strings.TrimRight(d.String(), "\n"), "body", body)}
}

func main() {
	brief := "no package"
	axis := "where a package sits in the handle space. " + brief +
		", read from the process, to see whether the package moves the generic and the process variable that come after the signals."
	kase("t81_pkt_none____", brief, "", "t80_stv_none____", nil, nil)

	one := pack("pk.pkg.vhdl", "pk", "a package with one integer constant", axis, 1)
	kase("t81_pkt_1con____", "a package with one integer constant", " + work.pk.c0",
		"t81_pkt_none____", []c.File{one}, []*c.Obj{con("pk", "c0", "integer", "1")})

	four := pack("pk.pkg.vhdl", "pk", "a package with four integer constants", axis, 4)
	kase("t81_pkt_4con____", "a package with four integer constants", " + work.pk.c0",
		"t81_pkt_1con____", []c.File{four},
		[]*c.Obj{con("pk", "c0", "integer", "1"), con("pk", "c1", "integer", "2"),
			con("pk", "c2", "integer", "3"), con("pk", "c3", "integer", "4")})

	fn := pack("pk.pkg.vhdl", "pk", "a package with one function", axis, 0)
	kase("t81_pkt_fn______", "a package with one function and no object", " + work.pk.f",
		"t81_pkt_1con____", []c.File{fn}, nil)

	// A package whose object is one array of sixteen integers, to see
	// whether the part past the second region moves with the object's
	// size.
	arr := c.File{Name: "pk.pkg.vhdl", Body: c.Fill(pkg, "brief", "a package with a constant array",
		"axis", axis, "name", "pk", "decls",
		"    type arr_t is array (0 to 15) of integer;\n    constant t : arr_t := (others => 7);", "body", "")}
	kase("t81_pkt_arr16___", "a package with a constant array of sixteen integers", " + work.pk.t(0)",
		"t81_pkt_1con____", []c.File{arr},
		[]*c.Obj{con("pk", "t", "arr_t", "(7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7)")})

	two := []c.File{
		pack("pk.pkg.vhdl", "pk", "the first of two packages", axis, 1),
		pack("qk.pkg.vhdl", "qk", "the second of two packages", axis, 1),
	}
	kase("t81_pkt_two_pk__", "two packages with one constant each",
		" + work.pk.c0 + work.qk.c0", "t81_pkt_1con____", two,
		[]*c.Obj{con("pk", "c0", "integer", "1"), con("qk", "c0", "integer", "1")})
}
