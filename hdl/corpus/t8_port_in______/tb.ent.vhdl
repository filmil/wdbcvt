-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal connected to a child's input port.
--!
--! Axis: ports. One input port driven by a parent signal, to see whether a port is an object of its own, with its own handle and records.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    dut: entity work.child port map (a => x);

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
