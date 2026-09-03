# SPDX-License-Identifier: Apache-2.0

# BUILD file for the Potato release archive; see MODULE.bazel. The
# file set is the SOURCE_FILES list of the Makefile, reordered.

package(default_visibility = ["//visibility:public"])

licenses(["notice"])  # BSD-3-Clause

exports_files(["LICENSE"])

# The processor. xvhdl needs a package before its first use and an
# entity before its first instantiation, so the list is in dependency
# order: packages, leaf entities, then the levels above them.
filegroup(
    name = "src",
    # do not sort
    srcs = [
        "src/pp_types.vhd",
        "src/pp_constants.vhd",
        "src/pp_utilities.vhd",
        "src/pp_csr.vhd",
        "src/pp_alu.vhd",
        "src/pp_alu_mux.vhd",
        "src/pp_alu_control_unit.vhd",
        "src/pp_comparator.vhd",
        "src/pp_counter.vhd",
        "src/pp_csr_alu.vhd",
        "src/pp_imm_decoder.vhd",
        "src/pp_fetch.vhd",
        "src/pp_memory.vhd",
        "src/pp_register_file.vhd",
        "src/pp_writeback.vhd",
        "src/pp_control_unit.vhd",
        "src/pp_csr_unit.vhd",
        "src/pp_decode.vhd",
        "src/pp_execute.vhd",
        "src/pp_core.vhd",
        "src/pp_icache.vhd",
        "src/pp_wb_adapter.vhd",
        "src/pp_wb_arbiter.vhd",
        "src/pp_potato.vhd",
    ],
)

# The processor bench of the release. It loads both memories from hex
# files named by string generics and stops on the first tohost write.
filegroup(
    name = "tb_processor",
    srcs = ["testbenches/tb_processor.vhd"],
)

# The data memory image of the release: one zero word.
filegroup(
    name = "empty_dmem",
    srcs = ["empty_dmem.hex"],
)
