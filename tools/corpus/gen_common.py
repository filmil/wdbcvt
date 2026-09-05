#!/usr/bin/env python3
"""Shared helpers of the tier 11 and later generators."""
import json, os, sys

# Under `bazel run` the script executes from its runfiles, and the
# corpus it writes is in the source tree, which BUILD_WORKSPACE_DIRECTORY
# names. Outside Bazel the path back from this file is the same place.
WORKSPACE = os.environ.get("BUILD_WORKSPACE_DIRECTORY") or os.path.dirname(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
ROOT = os.path.join(WORKSPACE, "hdl", "corpus")
# The comment marker of the language the file is in. A VHDL file with a
# Verilog comment on its first line does not compile, and the tiers
# before 60 were fixed by hand after every run.
HDRS = {".vhdl": "--", ".vhd": "--"}
HDR = "// SPDX-License-Identifier: Apache-2.0\n"


def header(filename):
    mark = HDRS.get(os.path.splitext(filename)[1], "//")
    return "%s SPDX-License-Identifier: Apache-2.0\n" % mark

def sig(scope, name, typ, width=1, **kw):
    d = {"scope": scope, "name": name, "width": width, "type": typ}
    d.update(kw)
    return d

# The database names no type for a vector, memory or struct field declared
# without a typedef, and gives a scalar reg, wire or logic the predefined
# type logic (bit for a two state scalar). The truth keeps the source
# keyword in "declared" and puts the database name in "type".
NETS = ("wire", "uwire", "wand", "wor", "tri", "triand", "trior", "tri0", "tri1", "supply0", "supply1")

def norm(sg):
    typ = sg["type"]
    if typ in ("reg", "logic", "bit") + NETS:
        sg["declared"] = typ
        if sg["width"] > 1:
            sg["type"] = ""
        elif typ != "logic" and typ != "bit":
            sg["type"] = "logic"
    elif typ == "memory":
        sg["declared"] = "memory of " + sg["element_type"]
        sg["type"] = ""
        sg["element_type"] = ""
    for f in sg.get("fields", []):
        norm(f)

# A Verilog source (a .v file) initialises every declared variable from
# an implicit initial block, so a four state object records all X at
# time zero and then its initial value. A real is not four state. A
# SystemVerilog source does that only for an enum or string initializer;
# those cases list the X record themselves.
def with_x(files, signals, transitions):
    if not any(fn.endswith(".v") for fn, _ in files):
        return transitions
    out = []
    seen = set()
    for x in transitions:
        name = x["signal"]
        if name not in seen:
            seen.add(name)
            sg = [g for g in signals if g["name"] == name or g["scope"] + "." + g["name"] == name][0]
            if sg["type"] != "real" and not str(x["value"]).startswith("X") and not str(x["value"]).startswith("(X"):
                if sg.get("elements"):
                    v = "(" + ", ".join(["X" * sg["element_width"]] * sg["elements"]) + ")"
                else:
                    v = "X" * sg["width"]
                out.append(tr(0, name, v))
        out.append(x)
    return out

def tr(t, s, v):
    return {"time_ns": t, "signal": s, "value": v}

def bits(v, n):
    return format(v, "0%db" % n)

# A memory initialised element by element records one change per element,
# all at time zero, after the all X record: t11_v_mem8 has eight. desc
# puts m[0] at the right.
def mem_trs(w, n, at, val, desc=False):
    x, z = "X" * w, bits(0, w)
    out = [tr(0, "m", "(" + ", ".join([x] * n) + ")")]
    cur = [x] * n
    for i in range(n):
        cur[n - 1 - i if desc else i] = z
        out.append(tr(0, "m", "(" + ", ".join(cur) + ")"))
    cur[n - 1 - at if desc else at] = bits(val, w)
    out.append(tr(50, "m", "(" + ", ".join(cur) + ")"))
    return out


def emit(name, axis, differs, files, signals, transitions, end=100, extra=None, x=True):
    assert len(name) == 16, name
    d = os.path.join(ROOT, name)
    os.makedirs(d, exist_ok=True)
    srcs = []
    for fn, body in files:
        with open(os.path.join(d, fn), "w") as f:
            f.write(header(fn) + body)
        srcs.append(fn)
    with open(os.path.join(d, "BUILD.bazel"), "w") as f:
        f.write('# SPDX-License-Identifier: Apache-2.0\n\n'
                'load("//build:wdb_case.bzl", "wdb_case")\n\n'
                'package(default_visibility = ["//visibility:public"])\n\n'
                'wdb_case(\n    name = "%s",\n    srcs = [\n' % name)
        for s in srcs:
            f.write('        "%s",\n' % s)
        f.write('    ],\n)\n')
    for sg in signals:
        norm(sg)
    if x:
        transitions = with_x(files, signals, transitions)
    t = {"case": name, "axis": axis, "differs_from": differs,
         "end_time_ns": end, "signals": signals, "transitions": transitions}
    if extra:
        t.update(extra)
    with open(os.path.join(d, "truth.json"), "w") as f:
        json.dump(t, f, indent=2)
        f.write("\n")
    print(name)

TS = "`timescale 1ns / 1ps\n\n"

TS = "`timescale 1ns / 1ps\n\n"
