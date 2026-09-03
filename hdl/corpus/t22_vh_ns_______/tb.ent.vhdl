-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: VHDL precision 1 ns
--!
--! Axis: xelab --timeprecision_vhdl 1ns
--!
--! The source is the same in every tier 22 case but t22_dbg_sub_proc;
--! the xelab options in BUILD.bazel are the axis.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (k : integer := 7);
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal n : integer := 0;
    signal t : time := 0 ps;

    function inc(x : integer) return integer is
        variable v : integer := x;
    begin
        v := v + 1;
        return v;
    end function;
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        n <= inc(k);
        t <= 1500 fs;
        wait for 1500 fs;
        s <= '0';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
