// SPDX-License-Identifier: Apache-2.0

// Tier 57: what log_wave can name, one object at a time.
package main

import (
	"fmt"
	"path/filepath"
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
    type rec_t is record
        a : std_ulogic;
        n : integer;
    end record;
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(3 downto 0) := "0000";
    signal r : rec_t := ('0', 0);
    constant c : integer := 3;
    shared variable sv : integer := 1;
begin
    g: for i in 0 to 1 generate
        signal gs : std_ulogic := '0';
    begin
        gs <= s;
    end generate;
    p: process
        variable w : integer := 7;
    begin
        for k in 0 to 2 loop
            w := w + k;
        end loop;
        wait for 10 ns;
        s <= '1';
        v <= "0101";
        r <= ('1', 5);
        sv := 2;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
`

const tcl = `open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
%s
run -all
close_vcd
exit
`

// logged names the objects the truth expects records for.
type logged []string

func (l logged) has(p string) bool {
	for _, e := range l {
		if e == p {
			return true
		}
	}
	return false
}

func (l logged) lg(p string) []any {
	if l.has(p) {
		return nil
	}
	return []any{"logged", false}
}

func kase(name, brief, axis string, lines []string, lgd logged, differs string, xelab []string) {
	body := c.Fill(tb, "brief", brief, "axis", axis)
	signals := []*c.Obj{
		c.Sig("tb", "s", "std_ulogic", 1).With(lgd.lg("tb.s")...),
		c.Sig("tb", "v", "std_ulogic_vector", 4).With(lgd.lg("tb.v")...),
		c.O("scope", "tb", "name", "r", "type", "rec_t", "fields", []*c.Obj{
			c.O("name", "a", "width", 1, "type", "std_ulogic"),
			c.O("name", "n", "width", 32, "type", "integer")}).With(lgd.lg("tb.r")...),
		c.Sig("tb.g(0)", "gs", "std_ulogic", 1).With(lgd.lg("tb.g(0).gs")...),
		c.Sig("tb.g(1)", "gs", "std_ulogic", 1).With(lgd.lg("tb.g(1).gs")...),
	}
	var trs []*c.Obj
	if lgd.has("tb.s") {
		trs = append(trs, c.Tr(0, "s", "0"), c.Tr(10, "s", "1"))
	}
	if lgd.has("tb.v") {
		trs = append(trs, c.Tr(0, "v", "0000"), c.Tr(10, "v", "0101"))
	}
	if lgd.has("tb.r") {
		trs = append(trs, c.Tr(0, "r.a", "0"), c.Tr(0, "r.n", "0"), c.Tr(10, "r.a", "1"), c.Tr(10, "r.n", "5"))
	}
	for i := 0; i < 2; i++ {
		p := fmt.Sprintf("tb.g(%d).gs", i)
		if lgd.has(p) {
			trs = append(trs, c.Tr(0, p, "0"), c.Tr(10, p, "1"))
		}
	}
	// A variable of the design. Everything but a process variable is
	// declared unlogged when the script does not name it.
	vr := func(scope, nm, kind string, kv ...any) *c.Obj {
		d := c.O("scope", scope, "name", nm, "type", "integer", "kind", kind)
		for i := 0; i < len(kv); i += 2 {
			d.Set(kv[i].(string), kv[i+1])
		}
		if kind != "variable" && !lgd.has(scope+"."+nm) {
			d.Set("logged", false)
		}
		return d
	}
	variables := []*c.Obj{
		vr("tb", "c", "constant", "value", "3"),
		vr("tb", "sv", "variable"),
		vr("tb.g(0)", "i", "loop", "value", "0"),
		vr("tb.g(1)", "i", "loop", "value", "1"),
		vr("tb.p", "w", "variable"),
		vr("tb.p", "k", "loop"),
	}
	if hasArg(xelab, "all") {
		// The library packages -debug all lists, as t22_dbg_all names them.
		t22, err := c.LoadObj(filepath.Join(c.Root(), "t22_dbg_all_____", "truth.json"))
		if err != nil {
			panic(err)
		}
		for _, v := range t22.Get("variables").([]any) {
			o := v.(*c.Obj)
			if !strings.HasPrefix(o.Str("scope"), "tb") {
				variables = append(variables, o)
			}
		}
	}
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{{Name: "tb.ent.vhdl", Body: body}}, Signals: signals, Trs: trs, End: 20,
		Extra: c.O("variables", variables), NoX: true})
	c.WriteFile(name, "xsim.tcl", strings.Replace(tcl, "%s", strings.Join(lines, "\n"), 1))
	tail := "    ],\n    tcl = \"xsim.tcl\",\n"
	if xelab != nil {
		tail += "    xelab_args = [\n"
		for _, a := range xelab {
			tail += "        \"" + a + "\",\n"
		}
		tail += "    ],\n"
	}
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n", tail+")\n", 1)
	})
}

func hasArg(l []string, s string) bool {
	for _, e := range l {
		if e == s {
			return true
		}
	}
	return false
}

var sigs = logged{"tb.s", "tb.v", "tb.r", "tb.g(0).gs", "tb.g(1).gs"}
var all = append(append(logged{}, sigs...), "tb.c", "tb.g(0).i", "tb.g(1).i", "tb.p.k")

const briefFmt = "log_wave naming %s of a design with every kind of object"
const axisFmt = "logging. log_wave names %s, in a design with a scalar, a vector, a record, a constant, a shared variable, a generate with a signal, and a process with a variable and a loop, to see what the database logs."

// one writes a case whose script logs a single object. vcd names what
// log_vcd takes when that is not the object itself, and noVCD leaves the
// VCD empty.
func one(name, what, obj string, lgd logged, vcd string, noVCD bool, xelab []string) {
	lines := []string{"log_wave " + obj}
	if !noVCD {
		v := vcd
		if v == "" {
			v = obj
		}
		lines = append([]string{"log_vcd " + v}, lines...)
	}
	kase(name, fmt.Sprintf(briefFmt, what), fmt.Sprintf(axisFmt, what), lines, lgd, "t57_log_all_____", xelab)
}

func main() {
	kase("t57_log_all_____", fmt.Sprintf(briefFmt, "everything, -recursive *"), fmt.Sprintf(axisFmt, "everything with -recursive *"),
		[]string{"log_vcd [get_objects -r /* ]", "log_wave -recursive *"}, all, "t7_gen_for______", nil)
	kase("t57_log_none____", fmt.Sprintf(briefFmt, "nothing"), fmt.Sprintf(axisFmt, "nothing, the script has no log_wave"),
		nil, nil, "t57_log_all_____", nil)
	one("t57_log_var_____", "a process variable", "/tb/p/w", nil, "", false, nil)
	one("t57_log_var_all_", "a process variable under -debug all", "/tb/p/w", nil, "", false, []string{"-debug", "all"})
	one("t57_log_shv_____", "a shared variable", "/tb/sv", nil, "", false, nil)
	one("t57_log_con_____", "an architecture constant", "/tb/c", logged{"tb.c"}, "", false, nil)
	one("t57_log_loop____", "a loop index", "/tb/p/k", logged{"tb.p.k"}, "", false, nil)
	one("t57_log_slice___", "a slice of a vector", "{/tb/v[2:1]}", logged{"tb.v"}, "", false, nil)
	// log_vcd of one bit writes the whole vector to the VCD, so the VCD
	// here logs nothing, to stay comparable with the database.
	one("t57_log_bit_____", "one bit of a vector", "{/tb/v[3]}", nil, "", true, nil)
	one("t57_log_rec_fld_", "a field of a record", "/tb/r.n", nil, "", false, nil)
	one("t57_log_rec_____", "a record signal", "/tb/r", logged{"tb.r"}, "", false, nil)
	one("t57_log_gen_sig_", "a signal of one generate iteration", `{/tb/\g(1)\/gs}`, logged{"tb.g(1).gs"}, "", false, nil)
	one("t57_log_gen_idx_", "the index of one generate iteration", `{/tb/\g(1)\/i}`, logged{"tb.g(1).i"}, "", false, nil)
	one("t57_log_gen_it__", "one generate iteration scope", `"/tb/\\g(1)\\"`, logged{"tb.g(1).gs", "tb.g(1).i"},
		`[get_objects {/tb/\g(1)\/*}]`, false, nil)
	one("t57_log_gen_____", "the generate statement scope", "/tb/g", nil, "[get_objects /tb/g/*]", false, nil)
	one("t57_log_proc____", "the process scope", "/tb/p", logged{"tb.p.k"}, "[get_objects /tb/p/*]", false, nil)
	one("t57_log_top_____", "the top scope without -recursive", "/tb", logged{"tb.s", "tb.v", "tb.r", "tb.c"},
		"[get_objects /tb/*]", false, nil)
}
