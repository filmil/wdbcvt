// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef enum {IDLE, RUN, DONE} state_t;
    state_t s = state_t'(1);
    initial begin
        #50 s = DONE;
        #50 $finish;
    end
endmodule
