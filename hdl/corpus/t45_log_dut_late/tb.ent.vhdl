-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the child scope logged from 10 ns.
--!
--! Axis: logging. The child log of t45_log_dut issued after run 10 ns.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    dut: entity work.child;

    p: process
    begin
        wait for 5 ns;
        s <= '1';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        s <= '1';
        wait for 5 ns;
        std.env.stop;
    end process;
end architecture;
