-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a procedure with 16 integer variables
--!
--! Axis: the space below the first signal. a procedure with 16 integer variables, to see what lies under the handle `0x768` that every first signal has, and whether filling it moves the signal.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure w(v : integer) is
        variable a0 : integer := 0;
        variable a1 : integer := 0;
        variable a2 : integer := 0;
        variable a3 : integer := 0;
        variable a4 : integer := 0;
        variable a5 : integer := 0;
        variable a6 : integer := 0;
        variable a7 : integer := 0;
        variable a8 : integer := 0;
        variable a9 : integer := 0;
        variable a10 : integer := 0;
        variable a11 : integer := 0;
        variable a12 : integer := 0;
        variable a13 : integer := 0;
        variable a14 : integer := 0;
        variable a15 : integer := 0;
    begin
        a0 := v + 0;
        a1 := v + 1;
        a2 := v + 2;
        a3 := v + 3;
        a4 := v + 4;
        a5 := v + 5;
        a6 := v + 6;
        a7 := v + 7;
        a8 := v + 8;
        a9 := v + 9;
        a10 := v + 10;
        a11 := v + 11;
        a12 := v + 12;
        a13 := v + 13;
        a14 := v + 14;
        a15 := v + 15;
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        w(5);
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
