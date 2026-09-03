-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal, then two children each with a signal and a process variable.
--!
--! Axis: logged objects. Logged and unlogged objects alternate, to see how many marker entries there are and what each holds.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    d0: entity work.child;
    d1: entity work.child;

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
