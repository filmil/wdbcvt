-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer subtype with narrow bounds
--!
--! Axis: subtype small_t is integer range 0 to 7

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    subtype small_t is integer range 0 to 7;
    signal s : small_t := 0;
begin
    p: process
    begin
        wait for 50 ns;
        s <= 5;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
