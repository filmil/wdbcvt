// SPDX-License-Identifier: Apache-2.0

// Corpus case: a packed struct
//
// Axis: debugging. a packed struct beside a logic, under -debug all where every SystemVerilog case before tier 60 ran under typical, to see what the flag adds to the type table and the debug section.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    typedef struct packed { logic a; logic [2:0] b; } st_t;
    st_t st = '0;

    initial begin
        #50 s = 1'b1;
        st = '{1, 3'b011};
        #50 $finish;
    end
endmodule
