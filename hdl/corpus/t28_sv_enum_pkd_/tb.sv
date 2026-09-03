// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef enum logic [1:0] {IDLE, RUN, DONE} state_t;
    state_t s = RUN;
    initial begin
        #50 s = DONE;
        #50 $finish;
    end
endmodule
