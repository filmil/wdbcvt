-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one edge at 5 s, 5e12 ps.
--!
--! Axis: time width. A change at 5 s, to see whether a time past 2^40 is stored whole.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 5 sec;
        s <= '1';
        wait for 5 sec;
        std.env.stop;
    end process;
end architecture;
