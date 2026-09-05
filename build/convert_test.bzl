# SPDX-License-Identifier: Apache-2.0

"""A test that converts one waveform database to FST."""

load("@rules_shell//shell:sh_test.bzl", "sh_test")

def fst_conversion_test(name, sim, size = "small", **kwargs):
    """Declares a test that converts `wdb` to FST and nothing else.

    The test passes when `wdbcvt -fst` exits zero and writes a file, so
    it says one thing: this database converts. The value checks live in
    //pkg/fstout, which reads the output back through libfst and
    compares it against the case's truth file; this target is what
    fails first, and per database, when a change breaks conversion for
    one design.

    Args:
      name: the test name.
      sim: the label of the simulation target. It produces the `.wdb`
        and the `.vcd`, and the script takes the first of them.
      size: the test size, "small" for a corpus case.
      **kwargs: forwarded to the test rule, for tags and timeouts.
    """
    sh_test(
        name = name,
        srcs = ["//build:fst_convert_test.sh"],
        args = [
            "$(rootpath //cmd/wdbcvt)",
            "$(rootpaths %s)" % sim,
        ],
        data = [
            "//cmd/wdbcvt",
            sim,
        ],
        size = size,
        **kwargs
    )
