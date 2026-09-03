// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef enum {IDLE, RUN, DONE} state_t;
    parameter state_t S = RUN;
    logic s = 1'b0;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
