// SPDX-License-Identifier: Apache-2.0

// Corpus case: two classes with a handle each
//
// Axis: debugging. two classes with a handle each beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    class a_t; int f = 1; endclass
    class b_t; int g = 2; endclass
    a_t ha;
    b_t hb;

    initial begin
        #50 s = 1'b1;
        ha = new; hb = new;
        #50 $finish;
    end
endmodule
