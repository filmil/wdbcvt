// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    parameter K = 1'b1 ? 3 : 4;
    logic s = 1'b0;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
