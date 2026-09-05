# SPDX-License-Identifier: Apache-2.0

# BUILD file for the Ibex archive; see MODULE.bazel. The lists follow
# the `.core` files of the project, which FuseSoC reads and this
# repository does not: `ibex_top_tracing.core` for the core, and
# `sim_shared.core` plus `ibex_simple_system.core` for the system its
# example builds around it. Order matters to xvlog, so packages come
# before the modules that import them.

package(default_visibility = ["//visibility:public"])

licenses(["notice"])  # Apache-2.0

exports_files(["LICENSE"])

# The assertion macros every RTL file includes, and the packages the
# primitives need. These are headers: they reach the compiler through
# the include path, not the command line.
filegroup(
    name = "prim_hdrs",
    srcs = [
        "vendor/lowrisc_ip/ip/prim/rtl/prim_assert.sv",
        "vendor/lowrisc_ip/ip/prim/rtl/prim_assert_dummy_macros.svh",
        "vendor/lowrisc_ip/ip/prim/rtl/prim_assert_sec_cm.svh",
        "vendor/lowrisc_ip/ip/prim/rtl/prim_assert_standard_macros.svh",
        "vendor/lowrisc_ip/ip/prim/rtl/prim_assert_yosys_macros.svh",
        "vendor/lowrisc_ip/ip/prim/rtl/prim_flop_macros.sv",
        "vendor/lowrisc_ip/ip/prim/rtl/prim_util_memload.svh",
    ],
)

# The functional coverage macros the RTL includes. They expand to
# nothing outside a coverage build, and the files still have to be
# found.
filegroup(
    name = "dv_hdrs",
    srcs = ["vendor/lowrisc_ip/dv/sv/dv_utils/dv_fcov_macros.svh"],
)

# The generic implementations of the abstract primitives. An ASIC flow
# picks a technology library here; a simulation takes these.
filegroup(
    name = "prims",
    # Compilation order matters to xvlog; do not sort.
    # do not sort
    srcs = [
        "vendor/lowrisc_ip/ip/prim/rtl/prim_util_pkg.sv",
        "vendor/lowrisc_ip/ip/prim/rtl/prim_secded_pkg.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_pkg.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_ram_1p_pkg.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_ram_2p_pkg.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_buf.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_flop.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_clock_gating.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_ram_1p.sv",
        "vendor/lowrisc_ip/ip/prim_generic/rtl/prim_ram_2p.sv",
    ],
)

# The core itself, with the tracing wrapper the example instantiates.
filegroup(
    name = "rtl",
    # Compilation order matters to xvlog; do not sort.
    # do not sort
    srcs = [
        "rtl/ibex_cheriot_pkg.sv",
        "rtl/ibex_pkg.sv",
        "rtl/ibex_tracer_pkg.sv",
        "rtl/ibex_alu.sv",
        "rtl/ibex_branch_predict.sv",
        "rtl/ibex_compressed_decoder.sv",
        "rtl/ibex_controller.sv",
        "rtl/ibex_counter.sv",
        "rtl/ibex_csr.sv",
        "rtl/ibex_cs_registers.sv",
        "rtl/ibex_decoder.sv",
        "rtl/ibex_dummy_instr.sv",
        "rtl/ibex_ex_block.sv",
        "rtl/ibex_fetch_fifo.sv",
        "rtl/ibex_id_stage.sv",
        "rtl/ibex_if_stage.sv",
        "rtl/ibex_load_store_unit.sv",
        "rtl/ibex_multdiv_fast.sv",
        "rtl/ibex_multdiv_slow.sv",
        "rtl/ibex_pmp.sv",
        "rtl/ibex_prefetch_buffer.sv",
        "rtl/ibex_register_file_ff.sv",
        "rtl/ibex_wb_stage.sv",
        "rtl/ibex_core.sv",
        "rtl/ibex_top.sv",
        "rtl/ibex_tracer.sv",
        "rtl/ibex_top_tracing.sv",
    ],
)

# The system the example builds around the core: a bus, a memory, a
# timer and the control register block that ends the run.
filegroup(
    name = "system",
    # Compilation order matters to xvlog; do not sort.
    # do not sort
    srcs = [
        "shared/rtl/ram_1p.sv",
        "shared/rtl/ram_2p.sv",
        "shared/rtl/bus.sv",
        "shared/rtl/timer.sv",
        "shared/rtl/sim/simulator_ctrl.sv",
        "examples/simple_system/rtl/ibex_simple_system.sv",
    ],
)
