// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    int s = 0;
    initial begin
        #50 for (int i = 0; i < 3; i++) s = s + i;
        #50 $finish;
    end
endmodule
