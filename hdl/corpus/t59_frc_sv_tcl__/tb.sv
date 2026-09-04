// SPDX-License-Identifier: Apache-2.0

// Corpus case: add_force on a SystemVerilog logic
//
// Axis: forcing. A second initial block does nothing, and the script forces the logic to 0 before the run on a logic driven 1 at 10 ns and 0 at 20 ns by the first, to see what the database records of a value the source imposes.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    initial begin
        #10 s = 1'b1;
        #10 s = 1'b0;
        #10 $finish;
    end
    initial begin
        #5 s = s;
        #10 s = s;
    end
endmodule
