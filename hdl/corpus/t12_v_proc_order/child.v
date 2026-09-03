// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t12_v_proc_order.

`timescale 1ns / 1ps

module child;
    reg s = 1'b0;
    wire w;

    assign w = ~s;

    initial begin
        #50 s = 1'b1;
    end
endmodule
