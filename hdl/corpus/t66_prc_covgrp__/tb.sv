// SPDX-License-Identifier: Apache-2.0

// Corpus case: a covergroup sampled from an always block
//
// Axis: process scopes. a covergroup sampled from an always block beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    covergroup cg @(posedge s);
        coverpoint s;
    endgroup
    cg c1 = new;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
