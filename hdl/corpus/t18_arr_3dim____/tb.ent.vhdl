-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a three dimensional array type
--!
--! Axis: array (0 to 1, 0 to 1, 0 to 2) of std_ulogic

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type cube_t is array (0 to 1, 0 to 1, 0 to 2) of std_ulogic;
    signal s : cube_t := (others => (others => (others => '0')));
begin
    p: process
    begin
        wait for 50 ns;
        s <= ((('1', '0', '1'), ('0', '1', '1')), (('1', '1', '0'), ('0', '0', '1')));
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
