// SPDX-License-Identifier: Apache-2.0

// Tier 59: forced and deposited values from the script.
package main

import (
	"fmt"
	"sort"
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
    signal v : std_ulogic_vector(3 downto 0) := "0000";
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        v <= "0101";
        wait for 10 ns;
        s <= '0';
        v <= "1010";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
`

const tcl = `open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
%s
close_vcd
exit
`

// rec is one record the truth expects: the time it lands at and the
// value it holds.
type rec struct {
	t int
	v string
}

// records turns the records a signal holds into its changes, and pins the
// count of records where a repeat of the value held makes the two
// differ, as tier 36 does.
func records(sg *c.Obj, recs []rec) []*c.Obj {
	var out []*c.Obj
	for _, r := range recs {
		if n := len(out); n > 0 && out[n-1].Str("value") == r.v {
			continue
		}
		out = append(out, c.Tr(r.t, sg.Str("name"), r.v))
	}
	if len(out) != len(recs) {
		sg.Set("records", len(recs))
	}
	return out
}

// driven is what the process writes when the script imposes nothing.
var drivenS = []rec{{0, "0"}, {10, "1"}, {20, "0"}}
var drivenV = []rec{{0, "0000"}, {10, "0101"}, {20, "1010"}}

// kase writes one VHDL case. s and v list the records the truth expects
// of the scalar and of the vector; nil means the driven values.
func kase(name, brief, axis string, lines []string, s, v []rec, differs string) {
	body := c.Fill(tb, "brief", brief, "axis", axis)
	sigs := []*c.Obj{c.Sig("tb", "s", "std_ulogic", 1), c.Sig("tb", "v", "std_ulogic_vector", 4)}
	if s == nil {
		s = drivenS
	}
	if v == nil {
		v = drivenV
	}
	trs := append(records(sigs[0], s), records(sigs[1], v)...)
	sort.SliceStable(trs, func(i, j int) bool { return trs[i].Int("time_ns") < trs[j].Int("time_ns") })
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{{Name: "tb.ent.vhdl", Body: body}}, Signals: sigs, Trs: trs, End: 30, NoX: true})
	writeTcl(name, lines)
}

func writeTcl(name string, lines []string) {
	c.WriteFile(name, "xsim.tcl", strings.Replace(tcl, "%s", strings.Join(lines, "\n"), 1))
	c.PatchBuild(name, func(b string) string {
		return strings.Replace(b, "    ],\n)\n", "    ],\n    tcl = \"xsim.tcl\",\n)\n", 1)
	})
}

const sv = `
// Corpus case: %(brief)s
//
// Axis: %(axis)s

` + "`" + `timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    initial begin
        #10 s = 1'b1;
        #10 s = 1'b0;
        #10 $finish;
    end
    initial begin
        #5 %(force)s;
        #10 %(release)s;
    end
endmodule
`

// svcase writes one SystemVerilog case, where the second initial block
// imposes the value rather than the script.
func svcase(name, brief, axis, force, release string, s []rec, differs string, tclLines []string) {
	body := c.Fill(sv, "brief", brief, "axis", axis, "force", force, "release", release)
	sigs := []*c.Obj{c.Sig("tb", "s", "logic", 1)}
	trs := records(sigs[0], s)
	c.Emit(c.Case{Name: name, Axis: axis, Differs: differs,
		Files: []c.File{{Name: "tb.sv", Body: body}}, Signals: sigs, Trs: trs, End: 30, NoX: true})
	if tclLines != nil {
		writeTcl(name, tclLines)
	}
}

const axisFmt = "forcing. The script %s, on a scalar driven 1 at 10 ns and 0 at 20 ns and a vector driven 0101 and 1010 at the same times, to see what the database records of a value the script imposes."
const svAxisFmt = "forcing. A second initial block %s on a logic driven 1 at 10 ns and 0 at 20 ns by the first, to see what the database records of a value the source imposes."

// pattern is the records of the scalar under a 0, 1 force repeating
// every 4 ns, beside the two the process writes.
func pattern() []rec {
	l := []rec{{10, "1"}, {20, "0"}}
	for t := 2; t < 29; t += 2 {
		v := "0"
		if (t/2)%2 == 1 {
			v = "1"
		}
		l = append(l, rec{t, v})
	}
	sort.Slice(l, func(i, j int) bool {
		if l[i].t != l[j].t {
			return l[i].t < l[j].t
		}
		return l[i].v < l[j].v
	})
	return append([]rec{{0, "0"}, {0, "0"}}, l...)
}

func main() {
	kase("t59_frc_none____", "the force design without a force",
		fmt.Sprintf(axisFmt, "forces nothing"), []string{"run -all"}, nil, nil, "t3_late_________")
	kase("t59_frc_s_const_", "a constant force on the scalar",
		fmt.Sprintf(axisFmt, "forces the scalar to 1 before the run"),
		[]string{"add_force /tb/s 1", "run -all"}, []rec{{0, "0"}, {0, "1"}, {20, "1"}}, nil, "t59_frc_none____")
	kase("t59_frc_s_cancel", "a force on the scalar cancelled after 5 ns",
		fmt.Sprintf(axisFmt, "forces the scalar to 1 and cancels the force after 5 ns"),
		[]string{"add_force /tb/s 1 -cancel_after 5ns", "run -all"},
		[]rec{{0, "0"}, {0, "1"}, {5, "0"}, {10, "1"}, {20, "0"}}, nil, "t59_frc_none____")
	kase("t59_frc_s_pat___", "a repeating force pattern on the scalar",
		fmt.Sprintf(axisFmt, "forces the scalar to a 0, 1 pattern every 4 ns"),
		[]string{"add_force /tb/s {0 0ns} {1 2ns} -repeat_every 4ns", "run -all"},
		pattern(), nil, "t59_frc_none____")
	kase("t59_frc_v_const_", "a constant force on the vector",
		fmt.Sprintf(axisFmt, "forces the vector to 1111 before the run"),
		[]string{"add_force /tb/v 1111", "run -all"}, nil,
		[]rec{{0, "0000"}, {0, "1111"}, {10, "1111"}, {20, "1111"}}, "t59_frc_none____")
	kase("t59_frc_v_bit___", "a force on one bit of the vector",
		fmt.Sprintf(axisFmt, "forces bit 3 of the vector to 1 before the run"),
		[]string{"add_force {/tb/v[3]} 1", "run -all"}, nil,
		[]rec{{0, "0000"}, {0, "1000"}, {10, "1101"}, {20, "1010"}}, "t59_frc_none____")
	kase("t59_frc_mid_____", "a force on the scalar at 15 ns",
		fmt.Sprintf(axisFmt, "runs 15 ns, then forces the scalar to 0"),
		[]string{"run 15 ns", "add_force /tb/s 0", "run -all"},
		[]rec{{0, "0"}, {10, "1"}, {15, "0"}}, nil, "t59_frc_none____")
	kase("t59_frc_release_", "a force on the scalar removed at 15 ns",
		fmt.Sprintf(axisFmt, "forces the scalar to 0, runs 15 ns, and removes the force"),
		[]string{"add_force /tb/s 0", "run 15 ns", "remove_forces /tb/s", "run -all"},
		[]rec{{0, "0"}, {0, "0"}, {10, "0"}, {15, "1"}, {15, "1"}, {20, "0"}}, nil, "t59_frc_none____")
	kase("t59_frc_deposit_", "a deposit on the scalar",
		fmt.Sprintf(axisFmt, "deposits 1 on the scalar with set_value before the run"),
		[]string{"set_value /tb/s 1", "run -all"}, []rec{{0, "0"}, {0, "1"}, {20, "0"}}, nil, "t59_frc_none____")
	kase("t59_frc_dep_mid_", "a deposit on the scalar at 15 ns",
		fmt.Sprintf(axisFmt, "runs 15 ns, then deposits 0 on the scalar with set_value"),
		[]string{"run 15 ns", "set_value /tb/s 0", "run -all"},
		[]rec{{0, "0"}, {10, "1"}, {15, "0"}}, nil, "t59_frc_none____")
	kase("t59_frc_dep_same", "a deposit of the value held",
		fmt.Sprintf(axisFmt, "deposits 0, the value held, on the scalar with set_value before the run"),
		[]string{"set_value /tb/s 0", "run -all"},
		[]rec{{0, "0"}, {0, "0"}, {10, "1"}, {20, "0"}}, nil, "t59_frc_none____")
	kase("t59_frc_mid_same", "a force of the value held at 15 ns",
		fmt.Sprintf(axisFmt, "runs 15 ns, then forces the scalar to 1, the value held"),
		[]string{"run 15 ns", "add_force /tb/s 1", "run -all"},
		[]rec{{0, "0"}, {10, "1"}, {15, "1"}, {20, "1"}}, nil, "t59_frc_none____")
	kase("t59_frc_rel_same", "a force removed while the driver agrees",
		fmt.Sprintf(axisFmt, "forces the scalar to 1, runs 15 ns, and removes the force while the driver holds 1"),
		[]string{"add_force /tb/s 1", "run 15 ns", "remove_forces /tb/s", "run -all"},
		[]rec{{0, "0"}, {0, "1"}, {15, "0"}, {15, "0"}}, nil, "t59_frc_none____")
	kase("t59_frc_twice___", "a second force at 15 ns",
		fmt.Sprintf(axisFmt, "forces the scalar to 1, runs 15 ns, and forces it to 0"),
		[]string{"add_force /tb/s 1", "run 15 ns", "add_force /tb/s 0", "run -all"},
		[]rec{{0, "0"}, {0, "1"}, {15, "0"}, {15, "0"}}, nil, "t59_frc_none____")

	svcase("t59_frc_sv_none_", "the SystemVerilog force design without a force",
		fmt.Sprintf(svAxisFmt, "does nothing"), "s = s", "s = s",
		[]rec{{0, "0"}, {10, "1"}, {20, "0"}}, "t59_frc_none____", nil)
	svcase("t59_frc_sv_force", "a force statement",
		fmt.Sprintf(svAxisFmt, "forces the logic to 1 at 5 ns and releases it at 15 ns"),
		"force s = 1'b1", "release s",
		[]rec{{0, "0"}, {5, "1"}, {20, "0"}}, "t59_frc_sv_none_", nil)
	svcase("t59_frc_sv_frc_0", "a force statement of 0",
		fmt.Sprintf(svAxisFmt, "forces the logic to 0 at 5 ns and releases it at 15 ns"),
		"force s = 1'b0", "release s",
		[]rec{{0, "0"}, {5, "0"}}, "t59_frc_sv_force", nil)
	svcase("t59_frc_sv_long_", "a force statement held over the second write",
		fmt.Sprintf(svAxisFmt, "forces the logic to 1 at 5 ns and releases it at 25 ns"),
		"force s = 1'b1", "#10 release s",
		[]rec{{0, "0"}, {5, "1"}}, "t59_frc_sv_force", nil)
	svcase("t59_frc_sv_norel", "a force statement without a release",
		fmt.Sprintf(svAxisFmt, "forces the logic to 1 at 5 ns and writes it to itself at 15 ns"),
		"force s = 1'b1", "s = s",
		[]rec{{0, "0"}, {5, "1"}}, "t59_frc_sv_long_", nil)
	svcase("t59_frc_sv_relon", "a release statement without a force",
		fmt.Sprintf(svAxisFmt, "writes the logic to itself at 5 ns and releases it at 15 ns"),
		"s = s", "release s",
		[]rec{{0, "0"}, {10, "1"}, {20, "0"}}, "t59_frc_sv_none_", nil)
	svcase("t59_frc_sv_tcl__", "add_force on a SystemVerilog logic",
		fmt.Sprintf(svAxisFmt, "does nothing, and the script forces the logic to 0 before the run"),
		"s = s", "s = s",
		[]rec{{0, "0"}, {10, "0"}}, "t59_frc_sv_none_", []string{"add_force /tb/s 0", "run -all"})
}
