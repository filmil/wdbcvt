// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    int s = 0;
    function int f();
        return 3;
    endfunction
    initial begin
        #50 s = f();
        #50 $finish;
    end
endmodule
