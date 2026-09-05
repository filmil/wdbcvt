// SPDX-License-Identifier: Apache-2.0

// Corpus case: a two bit enumeration between two logics
//
// Axis: an enumeration in a packed struct. a two bit enumeration between two logics, under typical, to see how wide the enumeration's bits are where a neighbour depends on it.

`timescale 1ns / 1ps

module tb;
    typedef enum logic [1:0] {E0 = 2'd0, E1 = 2'd1} e2_t;
    typedef struct packed { logic a; e2_t e; logic b; } rec_t;
    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};

    initial begin
        #50 s = '{a: 1'b1, e: E1, b: 1'b1};
        #50 $finish;
    end
endmodule
