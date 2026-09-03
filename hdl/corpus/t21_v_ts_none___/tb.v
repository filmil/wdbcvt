// SPDX-License-Identifier: Apache-2.0
// Corpus case: no timescale directive, Verilog.

module tb;
    reg s = 1'b0;
    initial begin
        #50 s = 1'b1;
        #50 $finish;
    end
endmodule
