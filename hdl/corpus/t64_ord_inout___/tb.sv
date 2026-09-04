// SPDX-License-Identifier: Apache-2.0

// Corpus case: a child inout on one bit beside a driver of another
//
// Axis: several partial drivers. a child inout on one bit beside a driver of another beside a logic, under typical, to see the order and the place of the records the drivers write.

`timescale 1ns / 1ps

module tb;
    logic s = 1'b0;
    wire [3:0] v; assign v[1] = s; bidi u(.io(v[3]));

    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule

module bidi(inout io);
    assign io = 1'b0;
endmodule
