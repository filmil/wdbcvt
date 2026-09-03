// SPDX-License-Identifier: Apache-2.0
// Corpus case: a clocking block.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic clk = 1'b0;
    always #25 clk = ~clk;
    clocking cb @(posedge clk);
        input s;
    endclocking

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
