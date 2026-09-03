-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a string signal with bounds 3 to 7.
--!
--! Axis: index range. The string signal of t44_str_sig with bounds (3 to 7), to see whether the declaration carries the bounds as written.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : string(3 to 7) := "hello";
begin
    p: process
    begin
        wait for 50 ns;
        s <= "world";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
