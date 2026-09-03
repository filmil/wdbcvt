// SPDX-License-Identifier: Apache-2.0
// Corpus case: an inout port, Verilog.

`timescale 1ns / 1ps

module tb;
    reg drv = 1'bz;
    wire w;

    assign w = drv;

    child dut(.w(w));

    initial begin
        #50 drv = 1'b1;
        #50 $finish;
    end
endmodule
