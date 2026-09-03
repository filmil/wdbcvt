// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t11_v_hier1.

`timescale 1ns / 1ps

module child;
    reg s = 1'b0;

    initial begin
        #50 s = 1'b1;
    end
endmodule
