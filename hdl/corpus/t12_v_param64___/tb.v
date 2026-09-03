// SPDX-License-Identifier: Apache-2.0
// Corpus case: a 64 bit parameter, Verilog.

`timescale 1ns / 1ps

module tb;
    child #(.W(64'hDEADBEEFCAFEBABE)) dut();

    initial begin
        #100 $finish;
    end
endmodule
