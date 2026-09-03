-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a parent signal, three open input ports and a child signal.
--!
--! Axis: ports. Three unconnected ports between two ordinary signals, to see the handle stride of a port that owns its handle.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    dut: entity work.child port map (a => open, b => open, c => open);

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
