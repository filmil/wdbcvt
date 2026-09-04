#!/usr/bin/env python3
"""Tier 57: what log_wave can name, one object at a time."""
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
"""

TCL = """open_vcd {{VCD_FILE}}
# Top entity: {{TOP}}
%s
run -all
close_vcd
exit
"""

SIGS = ["tb.s", "tb.v", "tb.r", "tb.g(0).gs", "tb.g(1).gs"]

def case(name, brief, axis, lines, logged, differs="t57_log_all_____", xelab=None):
    """logged names the objects the truth expects records for."""
    tb = TB % {"brief": brief, "axis": axis}
    def lg(p):
        return {} if p in logged else {"logged": False}
    signals = [
        dict(sig("tb", "s", "std_ulogic"), **lg("tb.s")),
        dict(sig("tb", "v", "std_ulogic_vector", 4), **lg("tb.v")),
        dict({"scope": "tb", "name": "r", "type": "rec_t", "fields": [
            {"name": "a", "width": 1, "type": "std_ulogic"},
            {"name": "n", "width": 32, "type": "integer"}]}, **lg("tb.r")),
        dict(sig("tb.g(0)", "gs", "std_ulogic"), **lg("tb.g(0).gs")),
        dict(sig("tb.g(1)", "gs", "std_ulogic"), **lg("tb.g(1).gs")),
    ]
    trs = []
    if "tb.s" in logged:
        trs += [tr(0, "s", "0"), tr(10, "s", "1")]
    if "tb.v" in logged:
        trs += [tr(0, "v", "0000"), tr(10, "v", "0101")]
    if "tb.r" in logged:
        trs += [tr(0, "r.a", "0"), tr(0, "r.n", "0"), tr(10, "r.a", "1"), tr(10, "r.n", "5")]
    for i in (0, 1):
        p = "tb.g(%d).gs" % i
        if p in logged:
            trs += [tr(0, p, "0"), tr(10, p, "1")]
    def var(scope, nm, kind, typ="integer", **kw):
        d = {"scope": scope, "name": nm, "type": typ, "kind": kind}
        d.update(kw)
        if kind != "variable" and scope + "." + nm not in logged:
            d["logged"] = False
        return d
    variables = [
        var("tb", "c", "constant", value="3"),
        var("tb", "sv", "variable"),
        var("tb.g(0)", "i", "loop", value="0"),
        var("tb.g(1)", "i", "loop", value="1"),
        var("tb.p", "w", "variable"),
        var("tb.p", "k", "loop"),
    ]
    if xelab and "all" in xelab:
        # The library packages -debug all lists, as t22_dbg_all names them.
        t22 = json.load(open(os.path.join(ROOT, "t22_dbg_all_____", "truth.json")))
        variables += [v for v in t22["variables"] if not v["scope"].startswith("tb")]
    emit(name, axis, differs, [("tb.ent.vhdl", tb)], signals, trs, end=20,
         extra={"variables": variables}, x=False)
    d = os.path.join(ROOT, name)
    with open(os.path.join(d, "xsim.tcl"), "w") as f:
        f.write(TCL % "\n".join(lines))
    p = os.path.join(d, "BUILD.bazel")
    b = open(p).read()
    tail = "    ],\n    tcl = \"xsim.tcl\",\n"
    if xelab:
        tail += "    xelab_args = [\n" + "".join('        "%s",\n' % a for a in xelab) + "    ],\n"
    b = b.replace("    ],\n)\n", tail + ")\n")
    open(p, "w").write(b)

ALL = SIGS + ["tb.c", "tb.g(0).i", "tb.g(1).i", "tb.p.k"]
BRIEF = "log_wave naming %s of a design with every kind of object"
AXIS = "logging. log_wave names %s, in a design with a scalar, a vector, a record, a constant, a shared variable, a generate with a signal, and a process with a variable and a loop, to see what the database logs."

def one(name, what, obj, logged, vcd=None, **kw):
    lines = ["log_wave %s" % obj]
    if vcd != "":
        lines.insert(0, "log_vcd %s" % (vcd or obj))
    case(name, BRIEF % what, AXIS % what, lines, logged, **kw)

if __name__ == "__main__":
    case("t57_log_all_____", BRIEF % "everything, -recursive *", AXIS % "everything with -recursive *",
         ["log_vcd [get_objects -r /* ]", "log_wave -recursive *"], ALL, differs="t7_gen_for______")
    case("t57_log_none____", BRIEF % "nothing", AXIS % "nothing, the script has no log_wave",
         [], [])
    one("t57_log_var_____", "a process variable", "/tb/p/w", [])
    one("t57_log_var_all_", "a process variable under -debug all", "/tb/p/w", [], xelab=["-debug", "all"])
    one("t57_log_shv_____", "a shared variable", "/tb/sv", [])
    one("t57_log_con_____", "an architecture constant", "/tb/c", ["tb.c"])
    one("t57_log_loop____", "a loop index", "/tb/p/k", ["tb.p.k"])
    one("t57_log_slice___", "a slice of a vector", "{/tb/v[2:1]}", ["tb.v"])
    # log_vcd of one bit writes the whole vector to the VCD, so the VCD
    # here logs nothing, to stay comparable with the database.
    one("t57_log_bit_____", "one bit of a vector", "{/tb/v[3]}", [], vcd="")
    one("t57_log_rec_fld_", "a field of a record", "/tb/r.n", [])
    one("t57_log_rec_____", "a record signal", "/tb/r", ["tb.r"])
    one("t57_log_gen_sig_", "a signal of one generate iteration", "{/tb/\\g(1)\\/gs}", ["tb.g(1).gs"])
    one("t57_log_gen_idx_", "the index of one generate iteration", "{/tb/\\g(1)\\/i}", ["tb.g(1).i"])
    one("t57_log_gen_it__", "one generate iteration scope", '"/tb/\\\\g(1)\\\\"', ["tb.g(1).gs", "tb.g(1).i"],
        vcd="[get_objects {/tb/\\g(1)\\/*}]")
    one("t57_log_gen_____", "the generate statement scope", "/tb/g", [], vcd="[get_objects /tb/g/*]")
    one("t57_log_proc____", "the process scope", "/tb/p", ["tb.p.k"], vcd="[get_objects /tb/p/*]")
    one("t57_log_top_____", "the top scope without -recursive", "/tb", ["tb.s", "tb.v", "tb.r", "tb.c"],
        vcd="[get_objects /tb/*]")
