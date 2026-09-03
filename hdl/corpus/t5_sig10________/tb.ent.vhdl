-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: ten one-bit signals, each with one edge.
--!
--! Axis: object count. More signals than fit one value page, to see how handles and pages are allocated.

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
        std.env.stop;
    end process;
end architecture;
