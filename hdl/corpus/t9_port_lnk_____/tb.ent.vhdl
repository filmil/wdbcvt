-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal connected to a child's linkage port.
--!
--! Axis: port mode. The linkage mode, to see whether declaration word 9 holds 4 for it.

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
