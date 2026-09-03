// SPDX-License-Identifier: Apache-2.0
// Corpus case: an associative array: int a[string].

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int a[string];

    initial begin
        #50 s = 1'b1;
        a["k"] = 5;
        #50 $finish;
    end
endmodule
