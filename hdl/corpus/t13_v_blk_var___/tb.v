// SPDX-License-Identifier: Apache-2.0
// Corpus case: a reg in a named block, Verilog.

`timescale 1ns / 1ps

module tb;
    reg s = 1'b0;

    initial begin : blk
        reg t;
        t = 1'b0;
        #50 t = 1'b1;
        s = t;
        #50 $finish;
    end
endmodule
