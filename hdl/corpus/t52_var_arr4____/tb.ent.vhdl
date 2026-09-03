-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two process variables, the first an integer array
--!
--! Axis: a process variable of an array of 4 integers before an integer one, for the handle stride

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type arr_t is array (0 to 3) of integer;
begin
    p: process
        variable a : arr_t := (others => 0);
        variable b : integer := 0;
    begin
        wait for 50 ns;
        b := b + 1;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
