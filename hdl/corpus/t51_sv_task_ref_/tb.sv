// SPDX-License-Identifier: Apache-2.0
// Corpus case: an automatic task with a ref argument

`timescale 1ns / 1ps

module tb;
    logic [7:0] s = 8'h00;

    task automatic inc(ref logic [7:0] v);
        begin
            v = v + 8'd1;
        end
    endtask
    initial begin
        #50 inc(s);
        #50 $finish;
    end
endmodule
