// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t12_v_port_reg.

`timescale 1ns / 1ps

module child(input a, output reg b);
    always @(a) b = ~a;
endmodule
