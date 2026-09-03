// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic [15:0] s = "a";
    initial begin
        #50 s = 16'ha5a5;
        #50 $finish;
    end
endmodule
