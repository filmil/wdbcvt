-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one bit, changes 1 ps and 1.5 ns apart.
--!
--! Axis: time resolution. Changes at 1 ps, 999 ps and 1000.5 ps, to see the unit and the rounding of record times.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 1 ps;
        s <= '1';
        wait for 998 ps;
        s <= '0';
        wait for 1500 fs;
        s <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
