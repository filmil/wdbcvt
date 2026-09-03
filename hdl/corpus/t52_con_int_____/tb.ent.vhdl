-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two architecture constants, the first of integer
--!
--! Axis: an architecture constant of integer before an integer one, for the handle stride

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    constant a : integer := 0;
    constant b : integer := 1;
begin
    p: process
    begin
        wait for 50 ns;

        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
