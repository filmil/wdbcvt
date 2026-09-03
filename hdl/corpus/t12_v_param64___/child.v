// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t12_v_param64.

`timescale 1ns / 1ps

module child #(parameter [63:0] W = 64'h0, parameter [7:0] P = 8'h5a);
    reg s = 1'b0;

    initial begin
        #50 s = 1'b1;
    end
endmodule
