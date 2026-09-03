// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int i = 5;
    initial begin
        #50 s = 1'b1;
        i = 165;
        #50 $finish;
    end
endmodule
