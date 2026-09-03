// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic [1:0] s [0:1] = '{default: 2'b01};
    initial begin
        #50 s[1] = 2'b11;
        #50 $finish;
    end
endmodule
