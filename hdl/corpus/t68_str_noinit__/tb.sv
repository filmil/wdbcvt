// SPDX-License-Identifier: Apache-2.0

// Corpus case: a string variable without an initializer
//
// Axis: string storage. a string variable without an initializer beside a logic, to see whether the characters of a SystemVerilog string reach the database, and where.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    string str;

    initial begin
        #50 s = 1'b1;
        str = "WPMK";
        #50 $finish;
    end
endmodule
