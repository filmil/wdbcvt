-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: only the child scope logged.
--!
--! Axis: logging. log_wave -recursive /tb/dut with a signal in tb as well, to see how a logged child under an unlogged parent is written.

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
