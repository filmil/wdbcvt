// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [31:0] m [0:3];

    initial begin
        $readmemb("hdl/corpus/t38_v_rmb_4w____/m.bin", m);
        #50 m[2] = 32'ha5c3f00f;
        #50 $finish;
    end
endmodule
