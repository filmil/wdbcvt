// SPDX-License-Identifier: Apache-2.0
// Corpus case: a static task with an output argument

`timescale 1ns / 1ps

module tb;
    logic [7:0] s = 8'h00;

    task inc(input logic [7:0] v, output logic [7:0] w);
        begin
            w = v + 8'd1;
        end
    endtask
    initial begin
        #50 inc(8'ha4, s);
        #50 $finish;
    end
endmodule
