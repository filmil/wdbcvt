// SPDX-License-Identifier: Apache-2.0

// Corpus case: the forty character string under -debug all
//
// Axis: string storage. the forty character string under -debug all beside a logic, to see whether the characters of a SystemVerilog string reach the database, and where.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    string str = "ZQXJWPMKZQXJWPMKZQXJWPMKZQXJWPMKZQXJWPMK";

    initial begin
        #50 s = 1'b1;
        str = "WPMK";
        #50 $finish;
    end
endmodule
