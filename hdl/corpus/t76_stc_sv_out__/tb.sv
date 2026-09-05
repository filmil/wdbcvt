// SPDX-License-Identifier: Apache-2.0

// Corpus case: an output argument of a task
//
// Axis: storage classes. an output argument of a task, to see which storage class word 28 of the instance record gives it, and whether any form gives the 5 that no case has produced.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int r = 0;
    task automatic two(output int y);
        y = 2;
    endtask

    initial begin
        #50 s = 1'b1;
        two(r);
        #50 $finish;
    end
endmodule
