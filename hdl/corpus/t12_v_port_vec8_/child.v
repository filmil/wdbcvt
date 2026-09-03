// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t12_v_port_vec8.

`timescale 1ns / 1ps

module child(input [7:0] a, output [7:0] b);
    assign b = ~a;
endmodule
