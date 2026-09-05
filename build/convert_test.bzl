# SPDX-License-Identifier: Apache-2.0

"""A test that converts one waveform database to every output."""

load("@rules_shell//shell:sh_test.bzl", "sh_test")

def conversion_test(name, sim, size = "small", sqlite = True, **kwargs):
    """Declares a test that converts `wdb` to FST and to SQLite.

    The test passes when `wdbcvt` exits zero and writes a file for each
    output, so it says one thing: this database converts. The value
    checks live in //pkg/fstout, which reads the FST back through
    libfst and compares it against the case's truth file, and in
    //pkg/sqlout, which compares its rows against Vivado's own VCD;
    this target is what fails first, and per database, when a change
    breaks conversion for one design.

    Args:
      name: the test name.
      sim: the label of the simulation target. It produces the `.wdb`
        and the `.vcd`, and the script takes the first of them.
      size: the test size, "small" for a corpus case.
      sqlite: whether to write the SQLite output too. A row per value
        change costs about 70 bytes on disk, which is fine for a corpus
        case and is a gigabyte for the largest designs, so those
        convert to FST here and are covered by //pkg/sqlout instead.
      **kwargs: forwarded to the test rule, for tags and timeouts.
    """
    outputs = "fst"
    if sqlite:
        outputs = "all"
    sh_test(
        name = name,
        srcs = ["//build:convert_test.sh"],
        args = [
            "$(rootpath //cmd/wdbcvt)",
            outputs,
            "$(rootpaths %s)" % sim,
        ],
        data = [
            "//cmd/wdbcvt",
            sim,
        ],
        size = size,
        **kwargs
    )
