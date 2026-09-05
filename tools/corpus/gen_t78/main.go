// SPDX-License-Identifier: Apache-2.0

// Tier 78: what a scope costs the handle space, and what a generate
// iteration saves against an instance of the same body.
package main

import (
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
begin
%(body)s
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
`

// solo is a child with no ports, so that an instance of it can be
// compared with a generate iteration that has no port either.
const solo = `
--! @file
--! @brief A child entity with no ports and one signal.

library ieee;
    use ieee.std_logic_1164.all;

entity solo is
end entity;

architecture sim of solo is
    signal g : std_ulogic := '0';
begin
    g <= '1' after 50 ns;
end architecture;
`

const kid = `
--! @file
--! @brief The child entity, whose architecture declares one signal.

library ieee;
    use ieee.std_logic_1164.all;

entity kid is
    port (
        i : in std_ulogic
    );
end entity;

architecture sim of kid is
    signal g : std_ulogic := '0';
begin
    g <= i;
end architecture;
`

func kase(name, brief, body, differs string, withKid bool, sigs, trs []*c.Obj, vars []*c.Obj) {
	kaseFiles(name, brief, body, differs, withKid, false, sigs, trs, vars)
}

func kaseFiles(name, brief, body, differs string, withKid, withSolo bool, sigs, trs []*c.Obj, vars []*c.Obj) {
	axis := "scope handle space. " + brief +
		", to see what a scope costs beyond its objects, and what a generate iteration saves against an instance of the same body."
	files := []c.File{}
	if withKid {
		files = append(files, c.File{Name: "kid.ent.vhdl", Body: c.Fill(kid)})
	}
	if withSolo {
		files = append(files, c.File{Name: "solo.ent.vhdl", Body: c.Fill(solo)})
	}
	files = append(files, c.File{Name: "tb.ent.vhdl",
		Body: c.Fill(tb, "brief", brief, "axis", axis, "body", body)})
	extra := c.O()
	if vars != nil {
		extra.Set("variables", vars)
	}
	if len(extra.Keys()) == 0 {
		extra = nil
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs, Files: files,
		Signals: append([]*c.Obj{c.Sig("tb", "s", "std_ulogic", 1)}, sigs...),
		Trs:     append([]*c.Obj{c.Tr(0, "s", "0"), c.Tr(50, "s", "1")}, trs...),
		Extra:   extra, NoX: true})
}

// gsig is the signal a generate iteration or an instance declares, and
// gtrs its records: it follows s.
func gsig(scope string) *c.Obj { return c.Sig(scope, "g", "std_ulogic", 1) }
func gtrs(scope string) []*c.Obj {
	return []*c.Obj{c.Tr(0, scope+".g", "0"), c.Tr(50, scope+".g", "1")}
}

func main() {
	// The baseline: one process and nothing else.
	kase("t78_scp_base____", "one process", "", "t52_var_int_____", false, nil, nil, nil)
	// Empty process scopes, which declare nothing.
	kase("t78_scp_proc2___", "a second process that declares nothing",
		`    q: process
    begin
        wait;
    end process;
`, "t78_scp_base____", false, nil, nil, nil)
	kase("t78_scp_proc3___", "two more processes that declare nothing",
		`    q: process
    begin
        wait;
    end process;

    r: process
    begin
        wait;
    end process;
`, "t78_scp_proc2___", false, nil, nil, nil)
	// A block, with and without a signal of its own.
	kase("t78_scp_block___", "a block that declares nothing",
		`    b: block
    begin
    end block;
`, "t78_scp_base____", false, nil, nil, nil)
	kase("t78_scp_blk_sig_", "a block with one signal",
		`    b: block
        signal g : std_ulogic := '0';
    begin
        g <= s;
    end block;
`, "t78_scp_block___", false, []*c.Obj{gsig("tb.b")}, gtrs("tb.b"), nil)
	// A generate iteration and an instance of a child, each declaring
	// the same one signal.
	kase("t78_scp_gen1____", "a generate of one iteration with a signal",
		`    g1: for i in 0 to 0 generate
        signal g : std_ulogic := '0';
    begin
        g <= s;
    end generate;
`, "t78_scp_blk_sig_", false, []*c.Obj{gsig("tb.g1(0)")}, gtrs("tb.g1(0)"),
		[]*c.Obj{c.O("scope", "tb.g1(0)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "0")})
	kase("t78_scp_inst1___", "an instance of a child with the same signal",
		"    u: entity work.kid port map (i => s);\n", "t78_scp_gen1____", true,
		[]*c.Obj{c.Sig("tb.u", "i", "std_ulogic", 1, "port", "in"), gsig("tb.u")},
		append([]*c.Obj{c.Tr(0, "tb.u.i", "0"), c.Tr(50, "tb.u.i", "1")}, gtrs("tb.u")...), nil)
	// Two of each, so that the per iteration and per instance costs
	// come out of the difference.
	kase("t78_scp_gen2____", "a generate of two iterations",
		`    g1: for i in 0 to 1 generate
        signal g : std_ulogic := '0';
    begin
        g <= s;
    end generate;
`, "t78_scp_gen1____", false,
		[]*c.Obj{gsig("tb.g1(0)"), gsig("tb.g1(1)")},
		append(gtrs("tb.g1(0)"), gtrs("tb.g1(1)")...),
		[]*c.Obj{
			c.O("scope", "tb.g1(0)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "0"),
			c.O("scope", "tb.g1(1)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "1")})
	// A child with no ports, which is the clean comparison: a generate
	// iteration has no port either.
	kaseFiles("t78_scp_solo1___", "one instance of a child with no ports",
		"    u: entity work.solo;\n", "t78_scp_gen1____", false, true,
		[]*c.Obj{gsig("tb.u")}, gtrs("tb.u"), nil)
	kaseFiles("t78_scp_solo2___", "two instances of a child with no ports",
		"    u: entity work.solo;\n    v: entity work.solo;\n", "t78_scp_solo1___", false, true,
		[]*c.Obj{gsig("tb.u"), gsig("tb.v")}, append(gtrs("tb.u"), gtrs("tb.v")...), nil)
	// A third copy of each, so that the fixed part and the per copy
	// part come out of three points rather than two.
	kase("t78_scp_gen3____", "a generate of three iterations",
		`    g1: for i in 0 to 2 generate
        signal g : std_ulogic := '0';
    begin
        g <= s;
    end generate;
`, "t78_scp_gen2____", false,
		[]*c.Obj{gsig("tb.g1(0)"), gsig("tb.g1(1)"), gsig("tb.g1(2)")},
		append(append(gtrs("tb.g1(0)"), gtrs("tb.g1(1)")...), gtrs("tb.g1(2)")...),
		[]*c.Obj{
			c.O("scope", "tb.g1(0)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "0"),
			c.O("scope", "tb.g1(1)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "1"),
			c.O("scope", "tb.g1(2)", "name", "i", "type", "integer", "kind", "loop", "port", "", "value", "2")})
	kaseFiles("t78_scp_solo3___", "three instances of a child with no ports",
		"    u: entity work.solo;\n    v: entity work.solo;\n    w: entity work.solo;\n",
		"t78_scp_solo2___", false, true,
		[]*c.Obj{gsig("tb.u"), gsig("tb.v"), gsig("tb.w")},
		append(append(gtrs("tb.u"), gtrs("tb.v")...), gtrs("tb.w")...), nil)
	kase("t78_scp_inst2___", "two instances of the child",
		"    u: entity work.kid port map (i => s);\n    v: entity work.kid port map (i => s);\n",
		"t78_scp_inst1___", true,
		[]*c.Obj{
			c.Sig("tb.u", "i", "std_ulogic", 1, "port", "in"), gsig("tb.u"),
			c.Sig("tb.v", "i", "std_ulogic", 1, "port", "in"), gsig("tb.v")},
		append(append([]*c.Obj{c.Tr(0, "tb.u.i", "0"), c.Tr(50, "tb.u.i", "1")}, gtrs("tb.u")...),
			append([]*c.Obj{c.Tr(0, "tb.v.i", "0"), c.Tr(50, "tb.v.i", "1")}, gtrs("tb.v")...)...), nil)
}
