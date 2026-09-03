// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic [63:0] s = 10ns;
    initial begin
        #50 s = 64'd165;
        #50 $finish;
    end
endmodule
