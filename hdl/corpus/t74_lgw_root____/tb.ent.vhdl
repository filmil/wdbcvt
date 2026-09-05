-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: log_wave -recursive /*
--!
--! Axis: the log_wave pattern. log_wave -recursive /*, to see why the default script's log_wave -recursive * leaves a package signal unlogged.

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
