// SPDX-License-Identifier: Apache-2.0
// Corpus case: a task with an input and a local reg, Verilog.

`timescale 1ns / 1ps

module tb;
    reg [7:0] s = 8'h00;

    task inc;
        input [7:0] v;
        reg [7:0] tmp;
        begin
            tmp = v + 8'd1;
            s = tmp;
        end
    endtask

    initial begin
        #50 inc(8'ha4);
        #50 $finish;
    end
endmodule
