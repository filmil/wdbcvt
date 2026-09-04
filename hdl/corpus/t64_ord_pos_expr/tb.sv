// SPDX-License-Identifier: Apache-2.0

// Corpus case: one child output on a scalar net, the input bound to an expression
//
// Axis: several partial drivers. one child output on a scalar net, the input bound to an expression beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire w; child u1(.i(~s), .o(w));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module child(input i, output o);
    assign o = i;
endmodule
