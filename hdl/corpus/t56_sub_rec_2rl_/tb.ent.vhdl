-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record local of two reals
--!
--! Axis: a record of two reals, one local of it, not initialised under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rec_t is record
        x : real;
        y : real;
    end record;
    function f(c : std_ulogic) return std_ulogic is
        variable r : rec_t;
        variable v : integer := 0;
    begin
        r.y := 1.0;
        v := v + integer(r.y);
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
