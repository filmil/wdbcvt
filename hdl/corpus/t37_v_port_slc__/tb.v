// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [7:0] r = 8'd0;
    wire [7:0] v = r;

    child u(.a(v[5:2]));

    initial begin
        #10 r = 8'hff;
        #10 r = 8'h3c;
        #10 r = 8'h4;
        #70 $finish;
    end
endmodule

module child(input [3:0] a);
    wire [3:0] i = a;
endmodule
