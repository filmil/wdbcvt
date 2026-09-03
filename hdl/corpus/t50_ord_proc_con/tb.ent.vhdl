-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a process constant declared before a variable
--!
--! Axis: a process with a constant above a variable, against t6_var_int with the variable alone

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
        constant c : integer := 3;
        variable v : integer := 0;
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
