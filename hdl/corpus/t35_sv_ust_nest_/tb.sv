// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef struct { logic a; logic [3:0] b; } rec_t;
    typedef struct { logic a; rec_t r; } outer_t;
    outer_t s;

    initial begin
        s = '{a: 1'b0, r: '{a: 1'b0, b: 4'h0}};
        #50 s.r.b = 4'ha;
        #50 $finish;
    end
endmodule
