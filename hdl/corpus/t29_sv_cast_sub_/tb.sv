// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    int s = 0;
    child u();
    initial begin
        #50 s = 165;
        #50 $finish;
    end
endmodule
