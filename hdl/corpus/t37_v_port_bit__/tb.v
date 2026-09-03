// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [7:0] r = 8'd0;
    wire [7:0] v = r;

    child u(.a(v[3]));

    initial begin
        #10 r = 8'h8;
        #10 r = 8'h0;
        #10 r = 8'hf7;
        #70 $finish;
    end
endmodule

module child(input a);
    wire i = a;
endmodule
