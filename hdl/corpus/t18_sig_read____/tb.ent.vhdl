-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the signal read by a concurrent assignment
--!
--! Axis: y <= s added beside the signal of t1_bit_one_edge

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal y : std_ulogic;
begin
    y <= s;

    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
