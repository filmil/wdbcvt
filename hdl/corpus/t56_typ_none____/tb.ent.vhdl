-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a function with a scalar local alone
--!
--! Axis: a function with an integer local, the baseline under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
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
    begin
        wait for 50 ns;
        s <= f('1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
