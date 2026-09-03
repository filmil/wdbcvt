-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two scalar constant locals
--!
--! Axis: a function with two local integer constants under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    function f(c : std_ulogic) return std_ulogic is
        constant k : integer := 3;
        constant k2 : integer := 4;
        variable v : integer := 0;
    begin
        v := v + k + k2;
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
