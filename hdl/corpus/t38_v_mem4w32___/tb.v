// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [31:0] m [0:3];

    initial begin
        m[0] = 32'h1; m[1] = 32'h2; m[2] = 32'h3; m[3] = 32'h4;
        #50 m[2] = 32'ha5c3f00f;
        #50 $finish;
    end
endmodule
