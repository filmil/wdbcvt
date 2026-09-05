-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record local of eight bytes, initialised
--!
--! Axis: static values in the handle space. a record local of eight bytes, initialised, to see whether the bytes of a static composite value push the objects that come after the signals.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        k : integer := 7
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rec_t is record
        m : integer;
        n : integer;
    end record;
    function f(c : std_ulogic) return std_ulogic is
        variable rc : rec_t := (0, 0);
        variable v : integer := 0;
    begin
        v := v + rc.m;
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
