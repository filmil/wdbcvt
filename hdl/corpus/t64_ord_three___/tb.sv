// SPDX-License-Identifier: Apache-2.0

// Corpus case: three instances of the child on three scalar nets
//
// Axis: several partial drivers. three instances of the child on three scalar nets beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire a, b, c; child u0(.i(s), .o(a)); child u1(.i(s), .o(b)); child u2(.i(s), .o(c));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module child(input i, output o);
    assign o = i;
endmodule
