-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array local assigned a dynamic aggregate
--!
--! Axis: a array local of four integers, not initialised, assigned an aggregate of a variable under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type arr_t is array (0 to 3) of integer;
    function f(c : std_ulogic) return std_ulogic is
        variable a : arr_t;
        variable v : integer := 0;
    begin
        v := v + 1;
        a := (others => v);
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
