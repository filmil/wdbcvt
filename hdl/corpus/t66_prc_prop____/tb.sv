// SPDX-License-Identifier: Apache-2.0

// Corpus case: a named property and sequence
//
// Axis: process scopes. a named property and sequence beside a logic, under typical, to see what scope the construct leaves and what it declares.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    logic c = 1'b0;
    always #10 c = ~c;
    sequence q; 1'b1; endsequence
    property p; @(posedge c) q; endproperty
    assert property (p);

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
