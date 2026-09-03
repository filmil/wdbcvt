// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    int s = 0;
    int a [0:2] = '{1, 2, 3};
    initial begin
        #50 foreach (a[i]) s = s + a[i];
        #50 $finish;
    end
endmodule
