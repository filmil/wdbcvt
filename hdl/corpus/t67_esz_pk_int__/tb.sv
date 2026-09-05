// SPDX-License-Identifier: Apache-2.0

// Corpus case: an enumeration over int in a packed struct
//
// Axis: an enumeration in a packed struct. an enumeration over int in a packed struct, under typical, to see how wide the enumeration's bits are where a neighbour depends on it.

`timescale 1ns / 1ps

module tb;
    typedef enum int {E0 = 0, E1 = 1} ei_t;
    typedef struct packed { logic a; ei_t e; logic b; } rec_t;
    rec_t s = '{a: 1'b0, e: E0, b: 1'b0};

    initial begin
        #50 s = '{a: 1'b1, e: E1, b: 1'b1};
        #50 $finish;
    end
endmodule
