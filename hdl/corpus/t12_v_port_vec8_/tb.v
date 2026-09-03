// SPDX-License-Identifier: Apache-2.0
// Corpus case: eight bit ports, Verilog.

`timescale 1ns / 1ps

module tb;
    reg [7:0] x = 8'h00;
    wire [7:0] y;

    child dut(.a(x), .b(y));

    initial begin
        #50 x = 8'ha5;
        #50 $finish;
    end
endmodule
