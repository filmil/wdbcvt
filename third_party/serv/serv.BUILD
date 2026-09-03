# SPDX-License-Identifier: Apache-2.0

# BUILD file for the SERV release archive; see MODULE.bazel. The file
# lists follow the `core` fileset of serv.core, the `rtl` fileset of
# servile.core, and the `soc` and `servant_tb` filesets of servant.core
# for a simulator that is not Quartus.

package(default_visibility = ["//visibility:public"])

licenses(["notice"])  # ISC

exports_files(["LICENSE"])

# The CPU.
filegroup(
    name = "rtl",
    srcs = [
        "rtl/serv_aligner.v",
        "rtl/serv_alu.v",
        "rtl/serv_bufreg.v",
        "rtl/serv_bufreg2.v",
        "rtl/serv_compdec.v",
        "rtl/serv_csr.v",
        "rtl/serv_ctrl.v",
        "rtl/serv_debug.v",
        "rtl/serv_decode.v",
        "rtl/serv_immdec.v",
        "rtl/serv_mem_if.v",
        "rtl/serv_rf_if.v",
        "rtl/serv_rf_ram.v",
        "rtl/serv_rf_ram_if.v",
        "rtl/serv_rf_top.v",
        "rtl/serv_state.v",
        "rtl/serv_top.v",
    ],
)

# The convenience wrapper around the CPU.
filegroup(
    name = "servile",
    srcs = [
        "servile/servile.v",
        "servile/servile_arbiter.v",
        "servile/servile_mux.v",
        "servile/servile_rf_mem_if.v",
    ],
)

# The reference SoC: timer, GPIO, RAM and the bus mux.
filegroup(
    name = "servant",
    srcs = [
        "servant/servant.v",
        "servant/servant_gpio.v",
        "servant/servant_mux.v",
        "servant/servant_ram.v",
        "servant/servant_timer.v",
    ],
)

# The simulation wrapper of the bench. Its UART decoder is left out:
# it declares `input rx` without a net type, and the `default_nettype
# none` of the files before it in one xvlog run makes that an error.
filegroup(
    name = "bench",
    srcs = ["bench/servant_sim.v"],
)

# The firmware image the RAM loads through $readmemh. A filegroup,
# because the simulation rule's data attribute takes targets only.
filegroup(
    name = "hello_uart",
    srcs = ["sw/hello_uart.hex"],
)
