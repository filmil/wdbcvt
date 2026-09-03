// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module child #(parameter real R = 1.5);
    integer s = 0;

    initial begin
        #50 s = 165;
    end
endmodule
