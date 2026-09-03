-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one edge at 5 ms, past 32 bits of picoseconds.
--!
--! Axis: time width. A change at 5 ms, 5e9 ps, to see whether a record time above 2^32 is stored whole.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 5 ms;
        s <= '1';
        wait for 5 ms;
        std.env.stop;
    end process;
end architecture;
