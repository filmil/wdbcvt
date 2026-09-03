-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an enumeration local
--!
--! Axis: a enumeration type declared in the architecture, one local of it under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type state_t is (idle, run, done);
    function f(c : std_ulogic) return std_ulogic is
        variable st : state_t := idle;
        variable v : integer := 0;
    begin
        v := v + 1;
        st := run;
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
