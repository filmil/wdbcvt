// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;
    always #25 clk = ~clk;
    reg s = 1'b0;

    always @(posedge clk) s <= 1'b0;

    child u(.a(s));

    initial begin
        #100 $finish;
    end
endmodule

module child(input a);
    wire i = a;
endmodule
