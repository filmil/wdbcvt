# SPDX-License-Identifier: Apache-2.0

# BUILD file for the NEORV32 release archive; see MODULE.bazel. The
# `core` list is `rtl/file_list_soc.f` of the release, in its order,
# which xvhdl needs; the `sim` list is the project's own testbench and
# its helpers, in the order they depend on each other.

package(default_visibility = ["//visibility:public"])

licenses(["notice"])  # BSD-3-Clause

exports_files(["LICENSE"])

# The processor, including the pre-initialized instruction memory image
# that boot mode 2 runs.
filegroup(
    name = "core",
    # Compilation order matters to xvhdl; do not sort.
    # do not sort
    srcs = [
        "rtl/core/neorv32_package.vhd",
        "rtl/core/neorv32_sys.vhd",
        "rtl/core/neorv32_fifo.vhd",
        "rtl/core/neorv32_cpu_decompressor.vhd",
        "rtl/core/neorv32_cpu_frontend.vhd",
        "rtl/core/neorv32_cpu_control.vhd",
        "rtl/core/neorv32_cpu_counters.vhd",
        "rtl/core/neorv32_cpu_regfile.vhd",
        "rtl/core/neorv32_cpu_cp_shifter.vhd",
        "rtl/core/neorv32_cpu_cp_muldiv.vhd",
        "rtl/core/neorv32_cpu_cp_bitmanip.vhd",
        "rtl/core/neorv32_cpu_cp_fpu.vhd",
        "rtl/core/neorv32_cpu_cp_cfu.vhd",
        "rtl/core/neorv32_cpu_cp_cond.vhd",
        "rtl/core/neorv32_cpu_cp_crypto.vhd",
        "rtl/core/neorv32_cpu_alu.vhd",
        "rtl/core/neorv32_cpu_lsu.vhd",
        "rtl/core/neorv32_cpu_pmp.vhd",
        "rtl/core/neorv32_cpu.vhd",
        "rtl/core/neorv32_cache.vhd",
        "rtl/core/neorv32_bus.vhd",
        "rtl/core/neorv32_dma.vhd",
        "rtl/core/neorv32_application_image.vhd",
        "rtl/core/neorv32_imem.vhd",
        "rtl/core/neorv32_dmem.vhd",
        "rtl/core/neorv32_xbus.vhd",
        "rtl/core/neorv32_bootloader_image.vhd",
        "rtl/core/neorv32_boot_rom.vhd",
        "rtl/core/neorv32_cfs.vhd",
        "rtl/core/neorv32_sdi.vhd",
        "rtl/core/neorv32_gpio.vhd",
        "rtl/core/neorv32_wdt.vhd",
        "rtl/core/neorv32_clint.vhd",
        "rtl/core/neorv32_uart.vhd",
        "rtl/core/neorv32_spi.vhd",
        "rtl/core/neorv32_twi.vhd",
        "rtl/core/neorv32_twd.vhd",
        "rtl/core/neorv32_pwm.vhd",
        "rtl/core/neorv32_trng.vhd",
        "rtl/core/neorv32_neoled.vhd",
        "rtl/core/neorv32_gptmr.vhd",
        "rtl/core/neorv32_onewire.vhd",
        "rtl/core/neorv32_slink.vhd",
        "rtl/core/neorv32_sysinfo.vhd",
        "rtl/core/neorv32_debug_dtm.vhd",
        "rtl/core/neorv32_debug_auth.vhd",
        "rtl/core/neorv32_debug_dm.vhd",
        "rtl/core/neorv32_top.vhd",
    ],
)

# The project's own testbench and the models it instantiates.
filegroup(
    name = "sim",
    # Compilation order matters to xvhdl; do not sort.
    # do not sort
    srcs = [
        "sim/sim_uart_rx.vhd",
        "sim/xbus_memory.vhd",
        "sim/xbus_gateway.vhd",
        "sim/neorv32_tb.vhd",
    ],
)
