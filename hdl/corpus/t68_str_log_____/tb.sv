// SPDX-License-Identifier: Apache-2.0

// Corpus case: the string named in log_wave under typical
//
// Axis: string storage. the string named in log_wave under typical beside a logic, to see whether the characters of a SystemVerilog string reach the database, and where.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    string str = "ZQXJ";

    initial begin
        #50 s = 1'b1;
        str = "WPMK";
        #50 $finish;
    end
endmodule
