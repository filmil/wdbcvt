-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a function nested in a function
--!
--! Axis: a function that declares a function of its own under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    function f(c : std_ulogic) return std_ulogic is
        variable v : integer := 0;
        function g(n : integer) return integer is
            variable w : integer := n;
        begin
            return w + 1;
        end function;
    begin
        v := g(2);
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
