-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a function with a variable, declared in the architecture and called from a process.
--!
--! Axis: subprograms. A function declaring a variable, to see whether a subprogram gets a scope, a unit or a declaration.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';

    function invert(v : std_ulogic) return std_ulogic is
        variable r : std_ulogic;
    begin
        r := not v;
        return r;
    end function;
begin
    p: process
    begin
        wait for 10 ns;
        x <= invert(x);
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
