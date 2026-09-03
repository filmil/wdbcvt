-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an element of a 2D array
--!
--! Axis: m(1, 0) <= '1' of an array (0 to 1, 0 to 2)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type mat_t is array (0 to 1, 0 to 2) of std_ulogic;
    signal m : mat_t := (others => (others => '0'));
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        m(1, 0) <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
