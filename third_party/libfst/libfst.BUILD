# SPDX-License-Identifier: Apache-2.0

# BUILD file for libfst, the FST reader and writer of GTKWave; see
# MODULE.bazel. Three source files and zlib. The upstream build
# generates a `config.h` with two probes and the parallel writer
# switch; `FST_INCLUDE_CONFIG` is left undefined instead, which skips
# the include, and the two probes are passed as defines.

load("@rules_cc//cc:defs.bzl", "cc_library")

package(default_visibility = ["//visibility:public"])

licenses(["notice"])  # MIT

exports_files(["LICENSE"])

cc_library(
    name = "fst",
    srcs = [
        "src/fastlz.c",
        "src/fastlz.h",
        "src/fstapi.c",
        "src/lz4.c",
        "src/lz4.h",
    ],
    hdrs = ["src/fstapi.h"],
    copts = [
        "-D_GNU_SOURCE",
        "-DHAVE_FSEEKO",
        "-DHAVE_REALPATH",
        "-w",
    ],
    includes = ["src"],
    # The host has no libstdc++ development files, and a cc_library
    # builds a shared object by default, whose link step wants them.
    linkstatic = True,
    deps = ["@zlib"],
)
