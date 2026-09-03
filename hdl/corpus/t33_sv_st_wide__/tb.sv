// SPDX-License-Identifier: Apache-2.0
// Corpus case: an unpacked struct with a 2400 bit field, the field written alone.

`timescale 1ns / 1ps

module tb;
    typedef struct { logic [2399:0] v; logic a; } st_t;
    st_t s = '{v: 2400'h0, a: 1'b0};

    initial begin
        #50 s.v = {75{32'ha5c3f00f}};
        #50 $finish;
    end
endmodule
