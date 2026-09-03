-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: one bit with one transition, one level down.
--!
--! Axis: hierarchy depth, 0 to 1. Isolates how a scope is opened and
--! closed, and how a signal is attached to one.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.child;

    p: process
    begin
        wait for 100 ns;
        std.env.stop;
    end process;
end architecture;
