// SPDX-License-Identifier: Apache-2.0

// Corpus case: two children of two modules on two scalar nets
//
// Axis: several partial drivers. two children of two modules on two scalar nets beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire a, b; child u0(.i(s), .o(a)); child2 u1(.i(s), .o(b));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module child(input i, output o);
    assign o = i;
endmodule

module child2(input i, output o);
    assign o = ~i;
endmodule
