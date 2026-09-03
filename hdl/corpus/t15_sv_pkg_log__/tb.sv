// SPDX-License-Identifier: Apache-2.0
// Corpus case: a package parameter under log_wave, SystemVerilog.

`timescale 1ns / 1ps

import p::*;

module tb;
    byte_t s = 8'h00;

    initial begin
        #50 s = W'(8'ha5);
        #50 $finish;
    end
endmodule
