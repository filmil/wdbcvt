// SPDX-License-Identifier: Apache-2.0

// Corpus case: two child outputs on two bits from a generate loop
//
// Axis: several partial drivers. two child outputs on two bits from a generate loop beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; genvar i; for (i = 0; i < 2; i = i + 1) begin : g child u(.i(s), .o(v[i * 3])); end

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module child(input i, output o);
    assign o = i;
endmodule
