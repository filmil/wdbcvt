-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one bit, 1300 transitions 1 ns apart.
--!
--! Axis: transition count. Enough value changes for three pages in one arena, to see where the marker record lands and how the arena record lists three pages.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        for i in 1 to 1300 loop
            wait for 1 ns;
            s <= not s;
        end loop;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
