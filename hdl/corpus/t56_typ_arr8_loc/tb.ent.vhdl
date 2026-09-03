-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array local of eight elements
--!
--! Axis: a array type of eight integers, one local of it under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type arr_t is array (0 to 7) of integer;
    function f(c : std_ulogic) return std_ulogic is
        variable a : arr_t := (others => 0);
        variable v : integer := 0;
    begin
        v := v + a(1);
        return c;
    end function;
begin
    p: process
    begin
        wait for 50 ns;
        s <= f('1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
