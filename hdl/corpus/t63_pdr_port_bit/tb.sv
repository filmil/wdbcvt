// SPDX-License-Identifier: Apache-2.0

// Corpus case: a child output bound to bit 1 of a 4 bit net
//
// Axis: partial drivers. a child output bound to bit 1 of a 4 bit net beside a logic, under typical, to see whether a driver of a bit, a slice or a port bound to part of a net records the whole net or the part.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; child u(.i(s), .o(v[1]));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module child(input i, output o);
    assign o = i;
endmodule
