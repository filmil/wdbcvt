-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a resolved signal connected to a child's inout port.
--!
--! Axis: ports. The port mode is inout, to fill in the mode word of the declaration record. The child drives Z so the parent's driver wins.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_logic := '0';
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
