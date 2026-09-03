// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [39:0] r = 40'd0;
    wire [39:0] v = r;

    child u(.a(v[35:28]));

    initial begin
        #10 r = 40'hff0000000;
        #10 r = 40'ha50000000;
        #10 r = 40'hfffffff;
        #70 $finish;
    end
endmodule

module child(input [7:0] a);
    wire [7:0] i = a;
endmodule
