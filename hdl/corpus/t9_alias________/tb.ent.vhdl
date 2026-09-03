-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal and an alias of it.
--!
--! Axis: alias. An alias of a signal, to see whether it is an object, and whether it shares the handle like a port.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
    alias y : std_ulogic is x;
begin
    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
