-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two array locals of four integers
--!
--! Axis: static values in the handle space. two array locals of four integers, to see whether the bytes of a static composite value push the objects that come after the signals.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        k : integer := 7
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type arr_t is array (0 to 3) of integer;
    function f(c : std_ulogic) return std_ulogic is
        variable ar : arr_t := (others => 0);
        variable br : arr_t := (others => 1);
        variable v : integer := 0;
    begin
        v := v + ar(1) + br(2);
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
