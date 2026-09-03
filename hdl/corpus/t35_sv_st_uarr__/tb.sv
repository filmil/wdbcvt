// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    typedef struct { logic a; logic [3:0] v [0:1]; } su_t;
    su_t s;

    initial begin
        s = '{a: 1'b0, v: '{4'h0, 4'h0}};
        #50 s.v[1] = 4'ha;
        #50 $finish;
    end
endmodule
