// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    wire w;
    assign w = s;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
