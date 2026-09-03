// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    child #(.K(7), .P(3)) dut();

    initial begin
        #100 $finish;
    end
endmodule
