// SPDX-License-Identifier: Apache-2.0
// Corpus case: the leaf module of t13_v_hier3_net.

`timescale 1ns / 1ps

module leaf(input i, output o);
    wire w2;
    reg r2 = 1'b0;
    assign w2 = i;
    assign o = w2;
endmodule
