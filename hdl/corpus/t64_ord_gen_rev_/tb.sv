// SPDX-License-Identifier: Apache-2.0

// Corpus case: the generate loop counting down
//
// Axis: several partial drivers. the generate loop counting down beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; genvar i; for (i = 3; i >= 0; i = i - 1) begin : g assign v[i] = s; end

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
