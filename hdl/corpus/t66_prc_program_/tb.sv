// SPDX-License-Identifier: Apache-2.0

// Corpus case: a program block instantiated beside the module
//
// Axis: process scopes. a program block instantiated beside the module beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    prog p();

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

program prog;
    initial begin
        #10;
    end
endprogram
