// SPDX-License-Identifier: Apache-2.0

// Corpus case: a driver of one whole element of an unpacked array of nets
//
// Axis: several partial drivers. a driver of one whole element of an unpacked array of nets beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v [0:1]; assign v[1] = {4{s}};

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
