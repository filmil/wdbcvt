// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    realtime s = 10ns;
    initial begin
        #50 s = 2.5;
        #50 $finish;
    end
endmodule
