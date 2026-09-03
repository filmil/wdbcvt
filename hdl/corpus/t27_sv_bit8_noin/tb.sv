// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    bit [7:0] s;
    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
