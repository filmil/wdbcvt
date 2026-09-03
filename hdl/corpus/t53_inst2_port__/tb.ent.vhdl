-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an input port of the child connected to s
--!
--! Axis: an input port of the child connected to s

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    d0: entity work.child generic map (k => 0) port map (a => s);
    d1: entity work.child generic map (k => 1) port map (a => s);

    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
