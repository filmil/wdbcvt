-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a child's output port driving a parent signal.
--!
--! Axis: ports. One output port driven inside the child, to see whether the port and the parent signal are one object or two.

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
