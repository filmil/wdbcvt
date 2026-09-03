-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a variable local that initialises another
--!
--! Axis: a function with a local variable read by the next initialiser under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    function f(c : std_ulogic) return std_ulogic is
        variable u : integer := 3;
        variable v : integer := u;
    begin
        v := v + 1;
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
