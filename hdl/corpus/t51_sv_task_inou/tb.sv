// SPDX-License-Identifier: Apache-2.0
// Corpus case: a static task with an inout argument

`timescale 1ns / 1ps

module tb;
    logic [7:0] s = 8'h00;

    task inc(inout logic [7:0] v);
        begin
            v = v + 8'd1;
        end
    endtask
    initial begin
        #50 inc(s);
        #50 $finish;
    end
endmodule
