// SPDX-License-Identifier: Apache-2.0

// Corpus case: an unpacked array of four bytes holding the same characters
//
// Axis: string storage. an unpacked array of four bytes holding the same characters beside a logic, to see whether the characters of a SystemVerilog string reach the database, and where.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    byte b [0:3] = '{"Z", "Q", "X", "J"};

    initial begin
        #50 s = 1'b1;
        b[0] = "W";
        #50 $finish;
    end
endmodule
