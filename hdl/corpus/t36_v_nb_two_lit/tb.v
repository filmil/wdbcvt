// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;

    initial begin
        #20 s <= 1'b0;
        #20 s <= 1'b0;
        #60 $finish;
    end
endmodule
