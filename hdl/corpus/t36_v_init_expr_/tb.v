// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg a = 1'b1;
    reg b = 1'b0;
    reg s = 1'b0;

    initial begin
        #50 s = a & b;
        #50 $finish;
    end
endmodule
