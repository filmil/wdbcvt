// SPDX-License-Identifier: Apache-2.0
// Corpus case: the interface of t15_sv_iface_mp.

`timescale 1ns / 1ps

interface bus_if;
    logic d;
    modport slave(input d);
endinterface
