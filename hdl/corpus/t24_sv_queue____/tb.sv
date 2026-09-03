// SPDX-License-Identifier: Apache-2.0
// Corpus case: a queue: int q[$].

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    int q[$];

    initial begin
        #50 s = 1'b1;
        q.push_back(5);
        #50 $finish;
    end
endmodule
