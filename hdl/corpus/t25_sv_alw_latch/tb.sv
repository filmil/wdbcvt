// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic en = 1'b0;
    logic d = 1'b0;
    logic q;
    always_latch if (en) q = d;
    initial begin
        #10 d = 1'b1;
        #10 en = 1'b1;
        #10 en = 1'b0;
        #10 d = 1'b0;
        #60 $finish;
    end
endmodule
