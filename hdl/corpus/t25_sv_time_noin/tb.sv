// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    time s;
    initial begin
        #50 s = 50;
        #50 $finish;
    end
endmodule
