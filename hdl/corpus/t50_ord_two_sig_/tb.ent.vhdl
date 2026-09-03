-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two signals declared out of alphabetical order
--!
--! Axis: signals z then a, to hold source order against name order

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal z : std_ulogic := '0';
    signal a : std_ulogic := '0';
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
