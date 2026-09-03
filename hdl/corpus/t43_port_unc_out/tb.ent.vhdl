-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an unconstrained output port
--!
--! Axis: ports. a : out std_ulogic_vector driven by the child, bound to an eight bit signal of tb.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal v : std_ulogic_vector(7 downto 0);
begin
    dut: entity work.child port map (a => v);
    p: process
    begin
        wait for 10 ns;
        null;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
