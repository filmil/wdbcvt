-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: 20 one-bit signals, each with one edge.
--!
--! Axis: object count. 20 signals, to see how the header arena table and the handle arenas grow.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s0 : std_ulogic := '0';
    signal s1 : std_ulogic := '0';
    signal s2 : std_ulogic := '0';
    signal s3 : std_ulogic := '0';
    signal s4 : std_ulogic := '0';
    signal s5 : std_ulogic := '0';
    signal s6 : std_ulogic := '0';
    signal s7 : std_ulogic := '0';
    signal s8 : std_ulogic := '0';
    signal s9 : std_ulogic := '0';
    signal s10 : std_ulogic := '0';
    signal s11 : std_ulogic := '0';
    signal s12 : std_ulogic := '0';
    signal s13 : std_ulogic := '0';
    signal s14 : std_ulogic := '0';
    signal s15 : std_ulogic := '0';
    signal s16 : std_ulogic := '0';
    signal s17 : std_ulogic := '0';
    signal s18 : std_ulogic := '0';
    signal s19 : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s0 <= '1';
        wait for 10 ns;
        s1 <= '1';
        wait for 10 ns;
        s2 <= '1';
        wait for 10 ns;
        s3 <= '1';
        wait for 10 ns;
        s4 <= '1';
        wait for 10 ns;
        s5 <= '1';
        wait for 10 ns;
        s6 <= '1';
        wait for 10 ns;
        s7 <= '1';
        wait for 10 ns;
        s8 <= '1';
        wait for 10 ns;
        s9 <= '1';
        wait for 10 ns;
        s10 <= '1';
        wait for 10 ns;
        s11 <= '1';
        wait for 10 ns;
        s12 <= '1';
        wait for 10 ns;
        s13 <= '1';
        wait for 10 ns;
        s14 <= '1';
        wait for 10 ns;
        s15 <= '1';
        wait for 10 ns;
        s16 <= '1';
        wait for 10 ns;
        s17 <= '1';
        wait for 10 ns;
        s18 <= '1';
        wait for 10 ns;
        s19 <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
