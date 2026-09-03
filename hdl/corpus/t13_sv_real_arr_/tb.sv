// SPDX-License-Identifier: Apache-2.0
// Corpus case: an unpacked array of reals, SystemVerilog.

`timescale 1ns / 1ps

module tb;
    real r [0:1];

    initial begin
        r[0] = 0.0;
        r[1] = 0.0;
        #50 r[1] = 1.5;
        #50 $finish;
    end
endmodule
