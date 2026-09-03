// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;
    integer i = 5;
    initial begin
        #50 s = 1'b1;
        i = 165;
        #50 $finish;
    end
endmodule
