// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    byte s;
    initial begin
        #50 s = 8'h05;
        #50 $finish;
    end
endmodule
