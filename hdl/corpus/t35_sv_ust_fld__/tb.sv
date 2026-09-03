// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef struct { logic a; logic [3:0] b; } rec_t;
    rec_t m [0:1];

    initial begin
        m[0] = '{a: 1'b0, b: 4'h0};
        m[1] = '{a: 1'b0, b: 4'h0};
        #50 m[1].b = 4'ha;
        #50 $finish;
    end
endmodule
