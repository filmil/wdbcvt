// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic [31:0] s = 32'h0;
    initial begin
        #50 s = 32'ha5;
        #50 $finish;
    end
endmodule
