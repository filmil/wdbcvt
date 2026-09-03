-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: 5 one-bit signals, each with one edge.
--!
--! Axis: object count. 5 signals, to see how the header arena table and the handle arenas grow.

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
        std.env.stop;
    end process;
end architecture;
