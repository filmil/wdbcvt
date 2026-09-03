// SPDX-License-Identifier: Apache-2.0
// Corpus case: three levels of wire and reg, Verilog.

`timescale 1ns / 1ps

module tb;
    wire w0;
    reg r0 = 1'b0;
    wire y;
    assign w0 = r0;
    mid dut(.i(w0), .o(y));

    initial begin
        #50 r0 = 1'b1;
        #50 $finish;
    end
endmodule
