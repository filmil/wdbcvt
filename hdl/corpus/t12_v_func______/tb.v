// SPDX-License-Identifier: Apache-2.0
// Corpus case: a function with an input and a local reg, Verilog.

`timescale 1ns / 1ps

module tb;
    reg [7:0] s = 8'h00;

    function [7:0] inc;
        input [7:0] v;
        reg [7:0] tmp;
        begin
            tmp = v + 8'd1;
            inc = tmp;
        end
    endfunction

    initial begin
        #50 s = inc(8'ha4);
        #50 $finish;
    end
endmodule
