// SPDX-License-Identifier: Apache-2.0
// Corpus case: log_wave naming the top without -recursive of a SystemVerilog design with every kind of object
//
// Axis: logging. log_wave names the top without -recursive, in a SystemVerilog design with a logic, a vector, a memory, a packed struct, an int, a real, a parameter, a localparam, a generate with a wire, a named block with a variable and a static task, to see what the database logs.

`timescale 1ns / 1ps

module tb;
    typedef struct packed { logic a; logic [3:0] b; } st_t;
    parameter P = 3;
    localparam L = 4;
    logic s = 1'b0;
    logic [3:0] v = 4'b0000;
    logic [3:0] m [0:1] = '{4'd0, 4'd0};
    st_t st = '{a: 1'b0, b: 4'b0000};
    int i = 7;
    real r = 0.5;
    genvar g;
    for (g = 0; g < 2; g++) begin : gb
        wire gw = s;
    end
    task inc(input int x);
        int tmp;
        tmp = x + 1;
        i = tmp;
    endtask
    initial begin : blk
        int bv;
        bv = 1;
        #10;
        s = 1'b1;
        v = 4'b0101;
        m[1] = 4'd9;
        st = '{a: 1'b1, b: 4'b0011};
        r = 1.5;
        inc(1);
        bv = 2;
        #10;
        $finish;
    end
endmodule
