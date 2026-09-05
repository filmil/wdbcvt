// SPDX-License-Identifier: Apache-2.0
//
// A bench around Ibex's own `simple_system`, which is the core with a
// bus, a memory, a timer and the control register block that ends the
// run. The example ships a C++ main for Verilator; this is the same
// system under a clock and a reset, with the program named by a
// parameter so the build can hand it the memory image.

module tb #(
    // The memory image the system boots from, given by the build. The
    // type is left off on purpose: xelab takes a string on the command
    // line for an untyped parameter, as //hdl/serv:sim does.
    parameter SRAM_INIT = "",
    // How long to run when the program does not stop the simulation
    // itself.
    parameter RUN_NS = 20000
);
    // Outside Verilator the system generates its own clock and reset,
    // so these two drive nothing. They are here because a port left
    // unconnected is X, and because the `ifdef in the system decides
    // which pair is used.
    logic clk = 1'b0;
    logic rst_n = 1'b0;

    always #1 clk = ~clk;

    initial begin
        #8 rst_n = 1'b1;
    end

    ibex_simple_system #(
        .SRAMInitFile(SRAM_INIT)
    ) u_system (
        .IO_CLK  (clk),
        .IO_RST_N(rst_n)
    );

    // The program ends the run by writing to the simulation control
    // register. This is the backstop for a program that does not.
    initial begin
        #(RUN_NS);
        $finish;
    end
endmodule
