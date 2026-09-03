// SPDX-License-Identifier: Apache-2.0
// Corpus case: handles. A wire with four continuous assignments declared before a reg, to read the cost of a fourth driver off the reg's handle.

`timescale 1ns / 1ps

module tb;
    wire w;
    reg s = 1'b0;
    assign w = s;
    assign w = s;
    assign w = s;
    assign w = s;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
