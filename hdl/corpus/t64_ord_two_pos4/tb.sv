// SPDX-License-Identifier: Apache-2.0

// Corpus case: two instances of a child with four ports
//
// Axis: several partial drivers. two instances of a child with four ports beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire c0, d0, c1, d1; child4 u0(.a(s), .b(s), .c(c0), .d(d0)); child4 u1(.a(s), .b(s), .c(c1), .d(d1));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module child4(input a, input b, output c, output d);
    assign c = a;
    assign d = b;
endmodule
