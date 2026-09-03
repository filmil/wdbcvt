// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    int s = 0;
    integer i;
    initial begin
        #50 for (i = 0; i < 3; i++) s = s + i;
        #50 $finish;
    end
endmodule
