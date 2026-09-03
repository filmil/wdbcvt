-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: three children each with a signal and a process variable.
--!
--! Axis: instance count. Three instances of a process with a variable, to see how many variable objects each process scope lists.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    d0: entity work.child;
    d1: entity work.child;
    d2: entity work.child;

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
