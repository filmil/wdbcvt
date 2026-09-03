// SPDX-License-Identifier: Apache-2.0
// Corpus case: the child module of t15_sv_iface_mp.

`timescale 1ns / 1ps

module child(bus_if.slave p);
    logic q;
    always_comb q = ~p.d;
endmodule
