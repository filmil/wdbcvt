// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [39:0] r = 40'd0;
    wire [39:0] v = r;

    child u(.a(v[39:34]));

    initial begin
        #10 r = 40'hff00000000;
        #10 r = 40'h5400000000;
        #10 r = 40'h3ffffffff;
        #70 $finish;
    end
endmodule

module child(input [5:0] a);
    wire [5:0] i = a;
endmodule
