// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    child #(.K(5)) dut();

    initial begin
        #100 $finish;
    end
endmodule
