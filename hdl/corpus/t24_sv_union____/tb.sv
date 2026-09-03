// SPDX-License-Identifier: Apache-2.0
// Corpus case: a packed union.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    typedef union packed {
        logic [7:0] b;
        logic [7:0] c;
    } u_t;
    u_t u = 8'h00;

    initial begin
        #50 s = 1'b1;
        u.b = 8'ha5;
        #50 $finish;
    end
endmodule
