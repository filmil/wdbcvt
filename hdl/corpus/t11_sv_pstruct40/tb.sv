// SPDX-License-Identifier: Apache-2.0
// Corpus case: a packed struct with a 40 bit field
//
// Differs from t11_sv_struct40 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    typedef struct packed { logic [39:0] a; logic b; } rec_t;
    rec_t s = '{a: 40'h0, b: 1'b0};

    initial begin
        #50 s = '{a: 40'ha5c3f00f12, b: 1'b1};
        #50 $finish;
    end
endmodule
