// SPDX-License-Identifier: Apache-2.0
// Corpus case: two instances with one parameter value, the child.

`timescale 1ns / 1ps

module child #(parameter K = 5);
    reg s = 1'b0;
    initial begin
        #50 s = 1'b1;
    end
endmodule
