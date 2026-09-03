-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a two dimensional array type
--!
--! Axis: array (0 to 1, 0 to 2) of std_ulogic, one type with two index ranges

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mat_t is array (0 to 1, 0 to 2) of std_ulogic;
    signal s : mat_t := (others => (others => '0'));
begin
    p: process
    begin
        wait for 50 ns;
        s <= (('1', '0', '1'), ('0', '1', '1'));
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
