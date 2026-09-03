// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg clk = 1'b0;
    always #25 clk = ~clk;
    reg [4:0] c = 5'd0;
    wire w;

    always @(posedge clk) c <= c + 5'd1;
    assign w = (c[4:2] == 3'b111);

    initial begin
        #100 $finish;
    end
endmodule
