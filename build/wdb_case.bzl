# SPDX-License-Identifier: Apache-2.0

"""A single corpus case: one simulation, one waveform database."""

load(
    "@rules_vivado//build/vivado:rules.bzl",
    "vivado_library",
    "vivado_simulation",
)

def wdb_case(name, srcs, extra_deps = []):
    """Declares one corpus case.

    Every case is held to the same shape on purpose, so that two cases
    differ only in their sources. The library name is always `corpus`
    and the top level entity is always `tb`, so neither can become a
    confound when two databases are compared. See //docs/corpus.md.

    Produces `<name>_sim.wdb` and `<name>_sim.vcd` under this package,
    and a `truth.json` filegroup that tests read as the ground truth for
    what the simulation did.

    Example:

        wdb_case(
            name = "t1_vec8",
            srcs = ["tb.ent.vhdl"],
        )

    Args:
      name: the case name. Matches the directory name.
      srcs: the VHDL sources, in compilation order.
      extra_deps: additional vivado_library targets to compile against.
    """
    native.filegroup(
        name = "srcs",
        # Compilation order matters to xvhdl.
        # do not sort
        srcs = srcs,
        tags = ["vhdl_ls"],
    )

    vivado_library(
        name = "lib",
        srcs = [":srcs"],
        library_name = "corpus",
        deps = extra_deps,
    )

    vivado_simulation(
        name = "sim",
        library = ":lib",
        top = "tb",
    )

    native.filegroup(
        name = "truth",
        srcs = ["truth.json"],
        visibility = ["//visibility:public"],
    )
