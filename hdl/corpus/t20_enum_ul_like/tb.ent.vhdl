-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an enumeration with the literals of std_ulogic
--!
--! Axis: the nine literals of STD_ULOGIC under another name

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type my9_t is ('U', 'X', '0', '1', 'Z', 'W', 'L', 'H', '-');
    signal s : my9_t := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
