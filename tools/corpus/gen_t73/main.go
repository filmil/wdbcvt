// SPDX-License-Identifier: Apache-2.0

// Tier 73: which scope the second pair of a protected type's method
// scopes hangs from.
package main

import (
	"strings"

	c "git.hdlfactory.com/HDL/wdbcvt/tools/corpus/gencommon"
)

// pkg declares the protected type, for the cases that put it in a
// package rather than in the architecture.
const pkg = `
--! @file
--! @brief Corpus case: %(brief)s
--!
--! Axis: %(axis)s

package pk is
    type counter_t is protected
        procedure bump;
        impure function get return integer;
    end protected;
end package;

package body pk is
    type counter_t is protected body
        variable n : integer := 0;
        procedure bump is
        begin
            n := n + 1;
        end procedure;
        impure function get return integer is
        begin
            return n;
        end function;
    end protected body;
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
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    %(decl)s
begin
    p: process
        variable v : integer := 0;
    begin
        wait for 50 ns;
        ct.bump;
        v := ct.get;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
%(tail)s
end architecture;
`

const kid = `
--! @file
--! @brief The child instantiated after the process.

library ieee;
    use ieee.std_logic_1164.all;

entity kid is
    port (
        i : in std_ulogic
    );
end entity;

architecture sim of kid is
begin
end architecture;
`

// inPackage is the declaration an architecture makes when the
// protected type lives in a package, and inArch declares the type
// itself in the architecture.
const inPackage = `shared variable ct : work.pk.counter_t;`

const inArch = `type counter_t is protected
        procedure bump;
        impure function get return integer;
    end protected;

    type counter_t is protected body
        variable n : integer := 0;
        procedure bump is
        begin
            n := n + 1;
        end procedure;
        impure function get return integer is
        begin
            return n;
        end function;
    end protected body;

    shared variable ct : counter_t;`

type opt struct {
	sigs      []*c.Obj
	trs       []*c.Obj
	variables []*c.Obj
	// arch declares the protected type in the architecture rather than
	// in a package, and kid adds the child entity the tail
	// instantiates.
	arch bool
	kid  bool
}

func kase(name, brief, tail, differs string, o opt) {
	axis := "protected method scopes. A shared variable of a protected type, its methods called from the only process, with " + brief +
		" after that process, to see which scope the second pair of method scopes hangs from."
	decl := inPackage
	if o.arch {
		decl = inArch
	}
	var files []c.File
	if !o.arch {
		files = append(files, c.File{Name: "pk.pkg.vhdl", Body: c.Fill(pkg, "brief", brief, "axis", axis)})
	}
	if o.kid {
		files = append(files, c.File{Name: "kid.ent.vhdl", Body: c.Fill(kid)})
	}
	files = append(files, c.File{Name: "tb.ent.vhdl",
		Body: c.Fill(tb, "brief", brief, "axis", axis, "decl", decl, "tail", tail)})
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs, Files: files,
		Signals: append([]*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)}, o.sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, o.trs...),
		Extra:   c.O("variables", o.variables), NoX: true})
	c.PatchBuild(name, func(b string) string {
		b = strings.Replace(b, "    srcs = [\n", "    # Compilation order matters to xvhdl; do not sort.\n    # do not sort\n    srcs = [\n", 1)
		return strings.Replace(b, "    ],\n)\n",
			"    ],\n    xelab_args = [\n        \"-debug\",\n        \"typical\",\n        \"-debug\",\n        \"subprogram\",\n    ],\n)\n", 1)
	})
}

// vars are the variables every case declares: the shared variable and
// the process's own.
func vars(extra ...*c.Obj) []*c.Obj {
	v := func(scope, name, typ string) *c.Obj {
		return c.O("scope", scope, "name", name, "type", typ, "kind", "variable", "port", "")
	}
	return append([]*c.Obj{v("tb", "ct", "counter_t"), v("tb.p", "v", "integer")}, extra...)
}

func main() {
	// A generate after the process, whose iterations are the last
	// scopes the writer visits.
	kase("t73_prt_gen_last", "a generate statement", `    g: for i in 0 to 1 generate
        signal gs : std_ulogic := '0';
    begin
        gs <= s;
    end generate;`, "t55_prot_pkg_prc",
		opt{sigs: []*c.Obj{c.Sig("tb.g(0)", "gs", "std_ulogic", 1), c.Sig("tb.g(1)", "gs", "std_ulogic", 1)},
			trs: []*c.Obj{c.Tr(0, "tb.g(0).gs", "0"), c.Tr(50, "tb.g(0).gs", "1"),
				c.Tr(0, "tb.g(1).gs", "0"), c.Tr(50, "tb.g(1).gs", "1")},
			variables: vars(
				c.O("scope", "tb.g(0)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "0"),
				c.O("scope", "tb.g(1)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "1"))})

	// A child instance after the process.
	kase("t73_prt_inst_lst", "an entity instance", "    u: entity work.kid port map (i => s);", "t73_prt_gen_last",
		opt{kid: true,
			sigs:      []*c.Obj{c.Sig("tb.u", "i", "std_ulogic", 1, "port", "in")},
			trs:       []*c.Obj{c.Tr(0, "tb.u.i", "0"), c.Tr(50, "tb.u.i", "1")},
			variables: vars()})

	// A block after the process, which is a scope with no process in
	// it at all.
	kase("t73_prt_blk_last", "a block statement", `    b: block
        signal bs : std_ulogic := '0';
    begin
        bs <= s;
    end block;`, "t73_prt_gen_last",
		opt{sigs: []*c.Obj{c.Sig("tb.b", "bs", "std_ulogic", 1)},
			trs:       []*c.Obj{c.Tr(0, "tb.b.bs", "0"), c.Tr(50, "tb.b.bs", "1")},
			variables: vars()})

	// The same generate with the type declared in the architecture,
	// where tier 55 saw the second pair under tb.
	kase("t73_prt_arch_gen", "a generate statement, with the type in the architecture", `    g: for i in 0 to 1 generate
        signal gs : std_ulogic := '0';
    begin
        gs <= s;
    end generate;`, "t73_prt_gen_last",
		opt{arch: true,
			sigs: []*c.Obj{c.Sig("tb.g(0)", "gs", "std_ulogic", 1), c.Sig("tb.g(1)", "gs", "std_ulogic", 1)},
			trs: []*c.Obj{c.Tr(0, "tb.g(0).gs", "0"), c.Tr(50, "tb.g(0).gs", "1"),
				c.Tr(0, "tb.g(1).gs", "0"), c.Tr(50, "tb.g(1).gs", "1")},
			variables: vars(
				c.O("scope", "tb.g(0)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "0"),
				c.O("scope", "tb.g(1)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "1"))})
}
