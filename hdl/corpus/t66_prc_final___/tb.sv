// SPDX-License-Identifier: Apache-2.0

// Corpus case: a final block
//
// Axis: process scopes. a final block beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    final begin
        $display("done");
    end

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
