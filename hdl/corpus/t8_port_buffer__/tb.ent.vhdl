-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a child's buffer port driving a parent signal.
--!
--! Axis: ports. The port mode is buffer, to fill in the mode word of the declaration record.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal y : std_ulogic;
begin
    dut: entity work.child port map (q => y);

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;
