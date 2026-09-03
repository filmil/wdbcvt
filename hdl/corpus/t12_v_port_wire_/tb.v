// SPDX-License-Identifier: Apache-2.0
// Corpus case: an input port driven by a wire, Verilog.

`timescale 1ns / 1ps

module tb;
    reg r = 1'b0;
    wire x;
    wire y;

    assign x = r;

    child dut(.a(x), .b(y));

    initial begin
        #50 r = 1'b1;
        #50 $finish;
    end
endmodule
