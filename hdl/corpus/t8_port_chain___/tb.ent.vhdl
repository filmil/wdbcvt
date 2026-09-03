-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal passed through two levels of input ports.
--!
--! Axis: ports. Three objects on one net, to see whether all three share the handle and how many records time 0 gets.

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
