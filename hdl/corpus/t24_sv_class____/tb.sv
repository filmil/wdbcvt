// SPDX-License-Identifier: Apache-2.0
// Corpus case: a class handle.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class c;
        int x = 3;
    endclass
    c h;

    initial begin
        #50 s = 1'b1;
        h = new;
        #50 $finish;
    end
endmodule
