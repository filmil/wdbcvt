// SPDX-License-Identifier: Apache-2.0

// Corpus case: a child bound into the module with bind
//
// Axis: process scopes. a child bound into the module with bind beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire w; assign w = s;

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module watcher(input i);
endmodule

bind tb watcher b(.i(s));
