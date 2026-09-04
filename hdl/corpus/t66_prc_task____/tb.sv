// SPDX-License-Identifier: Apache-2.0

// Corpus case: a task called from an initial block
//
// Axis: process scopes. a task called from an initial block beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    task t; #10; endtask
    initial t();

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
