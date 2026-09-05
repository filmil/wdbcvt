// SPDX-License-Identifier: Apache-2.0

// Corpus case: the array of two strings under -debug all
//
// Axis: string storage. the array of two strings under -debug all beside a logic, to see whether the characters of a SystemVerilog string reach the database, and where.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    string a [0:1] = '{"ZQXJ", "WPMK"};

    initial begin
        #50 s = 1'b1;
        a[1] = "ZQXJ";
        #50 $finish;
    end
endmodule
