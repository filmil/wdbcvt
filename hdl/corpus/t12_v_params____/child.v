// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t12_v_params.

`timescale 1ns / 1ps

module child #(parameter K = 5, parameter [7:0] P = 8'h5a, parameter integer Q = 9, parameter real R = 1.5);
    localparam L = K + 1;
    reg s = 1'b0;

    initial begin
        #50 s = 1'b1;
    end
endmodule
