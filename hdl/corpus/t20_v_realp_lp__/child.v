// SPDX-License-Identifier: Apache-2.0
// Corpus case: a real localparam, the child.

`timescale 1ns / 1ps

module child ;
    localparam real R = 1.5;
    reg s = 1'b0;
    initial begin
        #50 s = 1'b1;
    end
endmodule
