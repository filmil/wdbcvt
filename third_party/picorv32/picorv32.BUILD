# SPDX-License-Identifier: Apache-2.0

# BUILD file for the PicoRV32 archive; see MODULE.bazel. The core is one
# file, and `testbench_ez.v` is the bench of the project that needs no
# firmware image: it holds a six instruction program in an `initial`
# block.

package(default_visibility = ["//visibility:public"])

licenses(["unencumbered"])  # public domain

exports_files(["COPYING"])

# The CPU.
filegroup(
    name = "rtl",
    srcs = ["picorv32.v"],
)

# The bench that carries its own program.
filegroup(
    name = "bench",
    srcs = ["testbench_ez.v"],
)
