-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array type indexed -2 to 1
--!
--! Axis: index range. The range below zero in the type declaration, not the signal.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type int_array_t is array (-2 to 1) of integer;
    signal s : int_array_t := (others => 0);
begin
    p: process
    begin
        wait for 10 ns;
        s <= (-2 => 1, -1 => -2, 0 => 300, 1 => -40000);
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
