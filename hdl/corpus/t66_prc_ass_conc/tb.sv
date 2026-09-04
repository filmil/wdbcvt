// SPDX-License-Identifier: Apache-2.0

// Corpus case: a concurrent assertion on a clock
//
// Axis: process scopes. a concurrent assertion on a clock beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic c = 1'b0;
    always #10 c = ~c;
    assert property (@(posedge c) 1'b1);

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
