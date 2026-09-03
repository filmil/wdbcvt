-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one of two signals logged.
--!
--! Axis: logging. log_wave /tb/s with a second signal u declared, to see how an unlogged signal of the top is written: declared and marked, or absent.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal u : std_ulogic := '0';
begin
    p: process
    begin
        wait for 5 ns;
        s <= '1';
        u <= '1';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        s <= '1';
        wait for 5 ns;
        std.env.stop;
    end process;
end architecture;
