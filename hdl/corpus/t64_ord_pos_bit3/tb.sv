// SPDX-License-Identifier: Apache-2.0

// Corpus case: one child output on bit 3, the input bound to the logic
//
// Axis: several partial drivers. one child output on bit 3, the input bound to the logic beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; child u1(.i(s), .o(v[3]));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module child(input i, output o);
    assign o = i;
endmodule
