-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a function with a scalar local and no static value
--!
--! Axis: static values in the handle space. a function with a scalar local and no static value, to see whether the bytes of a static composite value push the objects that come after the signals.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        k : integer := 7
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    
    function f(c : std_ulogic) return std_ulogic is
        
        variable v : integer := 0;
    begin
        v := v + 1;
        return c;
    end function;
begin
    p: process
        variable a : integer := 0;
    begin
        wait for 50 ns;
        a := k;
        s <= f('1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
