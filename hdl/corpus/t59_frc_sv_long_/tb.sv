// SPDX-License-Identifier: Apache-2.0

// Corpus case: a force statement held over the second write
//
// Axis: forcing. A second initial block forces the logic to 1 at 5 ns and releases it at 25 ns on a logic driven 1 at 10 ns and 0 at 20 ns by the first, to see what the database records of a value the source imposes.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    initial begin
        #10 s = 1'b1;
        #10 s = 1'b0;
        #10 $finish;
    end
    initial begin
        #5 force s = 1'b1;
        #10 #10 release s;
    end
endmodule
