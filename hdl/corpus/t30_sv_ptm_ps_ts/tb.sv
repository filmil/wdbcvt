// SPDX-License-Identifier: Apache-2.0
`timescale 1ps / 1ps

module tb;
    parameter T = 10ns;
    logic s = 1'b0;
    initial begin
        #50000 s = 1'b1;
        #50000 $finish;
    end
endmodule
