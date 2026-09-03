-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an unconstrained port bound to an ascending actual
--!
--! Axis: ports. The actual is std_ulogic_vector(0 to 7), so the port's range is ascending too.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal v : std_ulogic_vector(0 to 7) := x"A5";
begin
    dut: entity work.child port map (a => v);
    p: process
    begin
        wait for 10 ns;
        v <= x"5A";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
