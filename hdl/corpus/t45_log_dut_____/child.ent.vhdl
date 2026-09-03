-- SPDX-License-Identifier: Apache-2.0

library ieee;
    use ieee.std_logic_1164.all;

--! @file
--! @brief An entity with one signal driven from a process.

--! The signal toggles on its own.
--!
--! ```
--! time  0 ns  5 ns  15 ns  25 ns
--! c     0     1     0      1
--! ```
entity child is
end entity;

architecture sim of child is
    signal c : std_ulogic := '0';
begin
    p: process
    begin
        wait for 5 ns;
        c <= '1';
        wait for 10 ns;
        c <= '0';
        wait for 10 ns;
        c <= '1';
        wait;
    end process;
end architecture;
