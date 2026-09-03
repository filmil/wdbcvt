// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module tb;
    reg [31:0] m [0:511];

    initial begin
        $readmemh("hdl/corpus/t38_v_mem512____/m.hex", m);
        #50 m[511] = 32'ha5c3f00f;
        #50 $finish;
    end
endmodule
