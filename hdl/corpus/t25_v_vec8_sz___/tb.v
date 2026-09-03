// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [7:0] s = 8'h00;
    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
