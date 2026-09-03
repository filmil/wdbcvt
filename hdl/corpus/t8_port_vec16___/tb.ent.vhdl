-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an open sixteen bit input port followed by a signal.
--!
--! Axis: ports. A sixteen byte port value, to see whether a port's handle cost grows with its size the way a signal's does.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
begin
    dut: entity work.child port map (a => open);

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;
