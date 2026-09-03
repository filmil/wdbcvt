// SPDX-License-Identifier: Apache-2.0
// Corpus case: an enum without an initializer
//
// Differs from t11_sv_enum by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    typedef enum {IDLE, RUN, DONE} state_t;
    state_t s;

    initial begin
        #50 s = DONE;
        #50 $finish;
    end
endmodule
