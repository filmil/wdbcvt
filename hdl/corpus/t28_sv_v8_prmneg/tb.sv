// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    parameter K = -1;
    logic [7:0] s = K;
    initial begin
        #50 s = 8'h00;
        #50 $finish;
    end
endmodule
