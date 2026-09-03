-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal declared in a package, driven from the architecture.
--!
--! Axis: package signal. A signal outside any entity, to see where the hierarchy puts it.

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
