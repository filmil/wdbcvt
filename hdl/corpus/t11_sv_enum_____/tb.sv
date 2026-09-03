// SPDX-License-Identifier: Apache-2.0
// Corpus case: one SystemVerilog enum
//
// Differs from t11_sv_logic by the declaration of s only.

`timescale 1ns / 1ps

module tb;
    typedef enum {IDLE, RUN, DONE} state_t;
    state_t s = IDLE;

    initial begin
        #50 s = DONE;
        #50 $finish;
    end
endmodule
