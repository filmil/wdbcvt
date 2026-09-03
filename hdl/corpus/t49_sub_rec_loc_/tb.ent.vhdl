-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record local of a function
--!
--! Axis: a function with a local variable of a record type under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rec_t is record
        a : std_ulogic;
        n : integer;
    end record;
    function f(c : std_ulogic) return std_ulogic is
        variable r : rec_t := (a => c, n => 1);
    begin
        return r.a;
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
