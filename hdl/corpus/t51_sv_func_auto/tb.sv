// SPDX-License-Identifier: Apache-2.0
// Corpus case: an automatic function with an argument and a local

`timescale 1ns / 1ps

module tb;
    logic [7:0] s = 8'h00;

    function automatic logic [7:0] inc(input logic [7:0] v);
        logic [7:0] tmp;
        begin
            tmp = v + 8'd1;
            return tmp;
        end
    endfunction
    initial begin
        #50 s = inc(8'ha4);
        #50 $finish;
    end
endmodule
