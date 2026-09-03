-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of 73 integers, 292 bytes.
--!
--! Axis: value size. 292 bytes of four byte elements, to see whether a chunk boundary falls between elements or inside one.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type int_array_t is array (0 to 72) of integer;
    signal s : int_array_t := (others => 0);
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => 7);
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
