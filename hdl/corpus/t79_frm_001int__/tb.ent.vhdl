-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a procedure with 1 integer variables
--!
--! Axis: the space below the first signal. a procedure with 1 integer variables, to see what lies under the handle `0x768` that every first signal has, and whether filling it moves the signal.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure w(v : integer) is
        variable a0 : integer := 0;
    begin
        a0 := v + 0;
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
