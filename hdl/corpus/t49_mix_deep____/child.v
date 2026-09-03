// SPDX-License-Identifier: Apache-2.0
`timescale 1ns / 1ps

module child(input a, output b);
    leaf u(.a(a), .b(b));
endmodule
