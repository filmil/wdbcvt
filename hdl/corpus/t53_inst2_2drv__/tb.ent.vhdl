-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a std_logic signal driven by two processes
--!
--! Axis: a std_logic signal driven by two processes

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    d0: entity work.child generic map (k => 0);
    d1: entity work.child generic map (k => 1);

    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
