// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef struct packed { logic a; logic [3:0] b; } s_t;
    typedef s_t arr_t [0:1];
    arr_t m;

    initial begin
        m[0] = '{a: 1'b0, b: 4'h0};
        m[1] = '{a: 1'b0, b: 4'h0};
        #50 m[1] = '{a: 1'b1, b: 4'ha};
        #50 $finish;
    end
endmodule
