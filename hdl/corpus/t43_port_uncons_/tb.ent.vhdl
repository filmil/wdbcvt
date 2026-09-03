-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an unconstrained input port
--!
--! Axis: ports. a : in std_ulogic_vector bound to an eight bit signal, against the constrained port of t8_port_vec8.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal v : std_ulogic_vector(7 downto 0) := x"A5";
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
