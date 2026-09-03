// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic [7:0] s = 8'h00;
    wire [7:0] w;
    assign w = s + 1;
    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
