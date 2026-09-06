-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record of eight integers
--!
--! Axis: the four bytes of a record. a record of eight integers, to see how much handle space the static value of a record adds beyond the bytes the record declares.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rec_t is record
        a1 : integer;
        a2 : integer;
        a3 : integer;
        a4 : integer;
        a5 : integer;
        a6 : integer;
        a7 : integer;
        a8 : integer;
    end record;
    function f(c : std_ulogic) return std_ulogic is
        variable r : rec_t := (0, 0, 0, 0, 0, 0, 0, 0);
        variable v : integer := 0;
    begin
        v := v + r.a1;
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
