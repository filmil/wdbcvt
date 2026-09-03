-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal declared in a package, driven from the architecture.
--!
--! Axis: the log_wave command. The same design as t9_pkg_sig, logged from the root, to see whether a package signal is logged.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    work.sig_pkg.g <= x;

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
