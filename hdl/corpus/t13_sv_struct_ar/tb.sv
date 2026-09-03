// SPDX-License-Identifier: Apache-2.0
// Corpus case: an unpacked array of packed structs, SystemVerilog.

`timescale 1ns / 1ps

module tb;
    typedef struct packed { logic a; logic [3:0] b; } s_t;
    s_t m [0:1];

    initial begin
        m[0] = '{a: 1'b0, b: 4'h0};
        m[1] = '{a: 1'b0, b: 4'h0};
        #50 m[1] = '{a: 1'b1, b: 4'ha};
        #50 $finish;
    end
endmodule
