// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    real s = 8'h05;
    initial begin
        #50 s = 2.5;
        #50 $finish;
    end
endmodule
