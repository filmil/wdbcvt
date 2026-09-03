// SPDX-License-Identifier: Apache-2.0
// Corpus case: the middle module of t13_v_hier3_net.

`timescale 1ns / 1ps

module mid(input i, output o);
    wire w1;
    reg r1 = 1'b0;
    assign w1 = i;
    leaf u(.i(w1), .o(o));
endmodule
