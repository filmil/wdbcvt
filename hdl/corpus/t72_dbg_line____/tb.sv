// SPDX-License-Identifier: Apache-2.0

// Corpus case: -debug wave -debug line
//
// Axis: debug modes. The same design, a logic beside a function with a local, elaborated with -debug wave -debug line, to see what the mode writes and which of the flag bytes of header words 14 and 15 it sets.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int r;

    function automatic int f(input int x);
        int tmp;
        tmp = x + 1;
        return tmp;
    endfunction

    initial begin
        #50 s = 1'b1;
        r = f(1);
        #50 $finish;
    end
endmodule
