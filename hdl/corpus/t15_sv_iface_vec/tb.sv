// SPDX-License-Identifier: Apache-2.0
// Corpus case: an interface carrying a vector, SystemVerilog.

`timescale 1ns / 1ps

module tb;
    bus_if b();
    child dut(b);

    initial begin
        b.d = 1'b0;
        b.v = 8'h00;
        #50 b.d = 1'b1;
        b.v = 8'ha5;
        #50 $finish;
    end
endmodule
