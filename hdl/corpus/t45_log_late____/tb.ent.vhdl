-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: logging started at 10 ns.
--!
--! Axis: logging. log_wave and log_vcd are issued after run 10 ns, to see what the database holds for the time before, and how the first record of a late logged signal is written.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
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
