-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a procedure with a signal parameter, declared in the architecture and called from a process.
--!
--! Axis: subprograms. A procedure that drives a signal through a parameter, to see whether the formal is an object or costs handle space.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';

    procedure flip(signal s : inout std_ulogic) is
        variable r : std_ulogic;
    begin
        r := not s;
        s <= r;
    end procedure;
begin
    p: process
    begin
        wait for 10 ns;
        flip(x);
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
