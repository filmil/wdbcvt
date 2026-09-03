// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef struct packed { logic a; logic [3:0] b; } rec_t;
    rec_t s = '{a: 1'b0, b: 4'b0000};
    initial begin
        #50 s = '{a: 1'b1, b: 4'b1010};
        #50 $finish;
    end
endmodule
