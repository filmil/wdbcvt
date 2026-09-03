-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer subtype with range -8 to 7
--!
--! Axis: integer bounds. A subtype whose lower bound is below zero, against 0 to 7.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    subtype small_t is integer range -8 to 7;
    signal s : small_t := -3;
begin
    p: process
    begin
        wait for 10 ns;
        s <= 7;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
