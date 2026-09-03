// SPDX-License-Identifier: Apache-2.0
// Corpus case: a real parameter of a long value, the child.

`timescale 1ns / 1ps

module child #(parameter real R = 123456.789);
    reg s = 1'b0;
    initial begin
        #50 s = 1'b1;
    end
endmodule
