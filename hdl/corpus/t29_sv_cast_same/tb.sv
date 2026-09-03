// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    int s = int'(1.5) + int'(2.5);
    initial begin
        #50 s = 165;
        #50 $finish;
    end
endmodule
