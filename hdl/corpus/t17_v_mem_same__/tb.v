// SPDX-License-Identifier: Apache-2.0
// Corpus case: a memory element rewritten with its value, Verilog.

`timescale 1ns / 1ps

module tb;
    reg [7:0] m [0:3];

    initial begin
        m[0] = 8'h00; m[1] = 8'h00; m[2] = 8'h00; m[3] = 8'h00;
        #50 m[2] = 8'h00;
        #50 $finish;
    end
endmodule
