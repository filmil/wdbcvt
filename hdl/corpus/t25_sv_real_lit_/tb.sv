// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    real s = 1.5;
    initial begin
        #50 s = 2.5;
        #50 $finish;
    end
endmodule
