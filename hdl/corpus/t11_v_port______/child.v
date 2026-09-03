// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t11_v_port.

`timescale 1ns / 1ps

module child(input a, output b);
    assign b = ~a;
endmodule
