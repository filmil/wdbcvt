#!/usr/bin/env python3
"""Tier 59: forced and deposited values from the script."""
import sys, os, json
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen_common import *

TB = """
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
"""

TCL = """open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
log_vcd [get_objects /tb/*]
log_wave -recursive *
%s
close_vcd
exit
"""

# The truth lists the changes, and pins the count of records, repeats of
# the value held included, through records, as tier 36 does.
def records(sg, recs):
    out = []
    for t, x in recs:
        if out and out[-1]["value"] == x:
            continue
        out.append(tr(t, sg["name"], x))
    if len(out) != len(recs):
        sg["records"] = len(recs)
    return out

DRIVEN = {"s": [(0, "0"), (10, "1"), (20, "0")],
          "v": [(0, "0000"), (10, "0101"), (20, "1010")]}

def case(name, brief, axis, lines, s=None, v=None, differs="t59_frc_none____"):
    """s and v list the (time, value) records the truth expects."""
    tb = TB % {"brief": brief, "axis": axis}
    signals = [sig("tb", "s", "std_ulogic"), sig("tb", "v", "std_ulogic_vector", 4)]
    trs = []
    for sg, recs in zip(signals, (s if s is not None else DRIVEN["s"], v if v is not None else DRIVEN["v"])):
        trs += records(sg, recs)
    trs.sort(key=lambda d: d["time_ns"])
    emit(name, axis, differs, [("tb.ent.vhdl", tb)], signals, trs, end=30, x=False)
    d = os.path.join(ROOT, name)
    with open(os.path.join(d, "xsim.tcl"), "w") as f:
        f.write(TCL % "\n".join(lines))
    p = os.path.join(d, "BUILD.bazel")
    b = open(p).read()
    b = b.replace("    ],\n)\n", "    ],\n    tcl = \"xsim.tcl\",\n)\n")
    open(p, "w").write(b)

AXIS = "forcing. The script %s, on a scalar driven 1 at 10 ns and 0 at 20 ns and a vector driven 0101 and 1010 at the same times, to see what the database records of a value the script imposes."

if __name__ == "__main__":
    case("t59_frc_none____", "the force design without a force",
         AXIS % "forces nothing", ["run -all"], differs="t3_late_________")
    case("t59_frc_s_const_", "a constant force on the scalar",
         AXIS % "forces the scalar to 1 before the run",
         ["add_force /tb/s 1", "run -all"], s=[(0, "0"), (0, "1"), (20, "1")])
    case("t59_frc_s_cancel", "a force on the scalar cancelled after 5 ns",
         AXIS % "forces the scalar to 1 and cancels the force after 5 ns",
         ["add_force /tb/s 1 -cancel_after 5ns", "run -all"],
         s=[(0, "0"), (0, "1"), (5, "0"), (10, "1"), (20, "0")])
    case("t59_frc_s_pat___", "a repeating force pattern on the scalar",
         AXIS % "forces the scalar to a 0, 1 pattern every 4 ns",
         ["add_force /tb/s {0 0ns} {1 2ns} -repeat_every 4ns", "run -all"],
         s=[(0, "0"), (0, "0")] + sorted([(t, "1" if (t // 2) % 2 else "0") for t in range(2, 29, 2)] + [(10, "1"), (20, "0")]))
    case("t59_frc_v_const_", "a constant force on the vector",
         AXIS % "forces the vector to 1111 before the run",
         ["add_force /tb/v 1111", "run -all"], v=[(0, "0000"), (0, "1111"), (10, "1111"), (20, "1111")])
    case("t59_frc_v_bit___", "a force on one bit of the vector",
         AXIS % "forces bit 3 of the vector to 1 before the run",
         ["add_force {/tb/v[3]} 1", "run -all"], v=[(0, "0000"), (0, "1000"), (10, "1101"), (20, "1010")])
    case("t59_frc_mid_____", "a force on the scalar at 15 ns",
         AXIS % "runs 15 ns, then forces the scalar to 0",
         ["run 15 ns", "add_force /tb/s 0", "run -all"], s=[(0, "0"), (10, "1"), (15, "0")])
    case("t59_frc_release_", "a force on the scalar removed at 15 ns",
         AXIS % "forces the scalar to 0, runs 15 ns, and removes the force",
         ["add_force /tb/s 0", "run 15 ns", "remove_forces /tb/s", "run -all"],
         s=[(0, "0"), (0, "0"), (10, "0"), (15, "1"), (15, "1"), (20, "0")])
    case("t59_frc_deposit_", "a deposit on the scalar",
         AXIS % "deposits 1 on the scalar with set_value before the run",
         ["set_value /tb/s 1", "run -all"], s=[(0, "0"), (0, "1"), (20, "0")])
    case("t59_frc_dep_mid_", "a deposit on the scalar at 15 ns",
         AXIS % "runs 15 ns, then deposits 0 on the scalar with set_value",
         ["run 15 ns", "set_value /tb/s 0", "run -all"], s=[(0, "0"), (10, "1"), (15, "0")])
    case("t59_frc_dep_same", "a deposit of the value held",
         AXIS % "deposits 0, the value held, on the scalar with set_value before the run",
         ["set_value /tb/s 0", "run -all"], s=[(0, "0"), (0, "0"), (10, "1"), (20, "0")])
    case("t59_frc_mid_same", "a force of the value held at 15 ns",
         AXIS % "runs 15 ns, then forces the scalar to 1, the value held",
         ["run 15 ns", "add_force /tb/s 1", "run -all"], s=[(0, "0"), (10, "1"), (15, "1"), (20, "1")])
    case("t59_frc_rel_same", "a force removed while the driver agrees",
         AXIS % "forces the scalar to 1, runs 15 ns, and removes the force while the driver holds 1",
         ["add_force /tb/s 1", "run 15 ns", "remove_forces /tb/s", "run -all"],
         s=[(0, "0"), (0, "1"), (15, "0"), (15, "0")])
    case("t59_frc_twice___", "a second force at 15 ns",
         AXIS % "forces the scalar to 1, runs 15 ns, and forces it to 0",
         ["add_force /tb/s 1", "run 15 ns", "add_force /tb/s 0", "run -all"],
         s=[(0, "0"), (0, "1"), (15, "0"), (15, "0")])

SV = """
// Corpus case: %(brief)s
//
// Axis: %(axis)s

`timescale 1ns / 1ps

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
"""

def svcase(name, brief, axis, force, release, s, differs, tcl=None):
    body = SV % {"brief": brief, "axis": axis, "force": force, "release": release}
    signals = [sig("tb", "s", "logic")]
    trs = records(signals[0], s)
    emit(name, axis, differs, [("tb.sv", body)], signals, trs, end=30, x=False)
    if tcl:
        d = os.path.join(ROOT, name)
        with open(os.path.join(d, "xsim.tcl"), "w") as f:
            f.write(TCL % "\n".join(tcl))
        p = os.path.join(d, "BUILD.bazel")
        b = open(p).read()
        b = b.replace("    ],\n)\n", "    ],\n    tcl = \"xsim.tcl\",\n)\n")
        open(p, "w").write(b)

SVAXIS = "forcing. A second initial block %s on a logic driven 1 at 10 ns and 0 at 20 ns by the first, to see what the database records of a value the source imposes."

if __name__ == "__main__":
    svcase("t59_frc_sv_none_", "the SystemVerilog force design without a force",
           SVAXIS % "does nothing", "s = s", "s = s",
           [(0, "0"), (10, "1"), (20, "0")], "t59_frc_none____")
    svcase("t59_frc_sv_force", "a force statement",
           SVAXIS % "forces the logic to 1 at 5 ns and releases it at 15 ns",
           "force s = 1'b1", "release s",
           [(0, "0"), (5, "1"), (20, "0")], "t59_frc_sv_none_")
    svcase("t59_frc_sv_frc_0", "a force statement of 0",
           SVAXIS % "forces the logic to 0 at 5 ns and releases it at 15 ns",
           "force s = 1'b0", "release s",
           [(0, "0"), (5, "0")], "t59_frc_sv_force")
    svcase("t59_frc_sv_long_", "a force statement held over the second write",
           SVAXIS % "forces the logic to 1 at 5 ns and releases it at 25 ns",
           "force s = 1'b1", "#10 release s",
           [(0, "0"), (5, "1")], "t59_frc_sv_force")
    svcase("t59_frc_sv_norel", "a force statement without a release",
           SVAXIS % "forces the logic to 1 at 5 ns and writes it to itself at 15 ns",
           "force s = 1'b1", "s = s",
           [(0, "0"), (5, "1")], "t59_frc_sv_long_")
    svcase("t59_frc_sv_relon", "a release statement without a force",
           SVAXIS % "writes the logic to itself at 5 ns and releases it at 15 ns",
           "s = s", "release s",
           [(0, "0"), (10, "1"), (20, "0")], "t59_frc_sv_none_")
    svcase("t59_frc_sv_tcl__", "add_force on a SystemVerilog logic",
           SVAXIS % "does nothing, and the script forces the logic to 0 before the run",
           "s = s", "s = s",
           [(0, "0"), (10, "0")], "t59_frc_sv_none_", tcl=["add_force /tb/s 0", "run -all"])
