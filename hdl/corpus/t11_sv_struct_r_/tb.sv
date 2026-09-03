// SPDX-License-Identifier: Apache-2.0
// Corpus case: an unpacked struct with a real field
//
// Differs from t11_sv_struct3 by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    typedef struct { real r; logic a; } rec_t;
    rec_t s = '{r: 0.0, a: 1'b0};

    initial begin
        #50 s = '{r: 1.5, a: 1'b1};
        #50 $finish;
    end
endmodule
