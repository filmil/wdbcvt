// SPDX-License-Identifier: Apache-2.0
// Corpus case: an interface instance, SystemVerilog.

`timescale 1ns / 1ps

module tb;
    bus_if b();
    child dut(b);

    initial begin
        b.d = 1'b0;
        #50 b.d = 1'b1;
        #50 $finish;
    end
endmodule
