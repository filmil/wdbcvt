-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a vector indexed -4 to 3
--!
--! Axis: index range. An ascending range from below zero.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type vec_t is array (integer range <>) of std_ulogic;
    signal s : vec_t(-4 to 3) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= x"18";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
