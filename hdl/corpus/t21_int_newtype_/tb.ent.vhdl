-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer type with its own bounds
--!
--! Axis: type small_t is range 0 to 7, a new type rather than a subtype

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type small_t is range 0 to 7;
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
