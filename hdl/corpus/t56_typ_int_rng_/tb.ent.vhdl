-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer local with a range
--!
--! Axis: a integer range 0 to 7 local under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    function f(c : std_ulogic) return std_ulogic is
        variable n : integer range 0 to 7 := 0;
        variable v : integer := 0;
    begin
        v := v + 1;
        n := n + 1;
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
