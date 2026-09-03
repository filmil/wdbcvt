-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer array element
--!
--! Axis: a(1) <= 5 of an array (0 to 3) of integer

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_arr_t is array (0 to 3) of integer;
    signal a : int_arr_t := (others => 0);
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        a(1) <= 5;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
