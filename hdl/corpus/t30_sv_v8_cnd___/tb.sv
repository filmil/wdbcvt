// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    logic [7:0] s = 1'b1 ? 8'd5 : 0;
    initial begin
        #50 s = 8'ha5;
        #50 $finish;
    end
endmodule
