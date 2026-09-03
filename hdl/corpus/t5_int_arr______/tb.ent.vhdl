-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of four integers.
--!
--! Axis: array element type. Elements wider than a byte say whether an array is elements back to back.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type int_array_t is array (0 to 3) of integer;
    signal s : int_array_t := (others => 0);
begin
    p: process
    begin
        wait for 50 ns;
        s <= (1, -2, 300, -40000);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
