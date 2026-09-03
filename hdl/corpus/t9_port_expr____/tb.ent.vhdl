-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an input port bound to a literal.
--!
--! Axis: port association. The actual is a literal, to see whether the port gets a handle of its own like an open port.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    dut: entity work.child port map (a => '1');

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
