// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    parameter K = 32'h5a;
    logic [7:0] s = K[7:0];
    initial begin
        #50 s = 8'h00;
        #50 $finish;
    end
endmodule
