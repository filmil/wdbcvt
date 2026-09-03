-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an input port bound to one element of a parent vector.
--!
--! Axis: port association. The actual is an element of a vector, not a whole signal, to see whether the port still shares a handle or gets its own.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic_vector(1 downto 0) := "00";
begin
    dut: entity work.child port map (a => x(0));

    p: process
    begin
        wait for 10 ns;
        x <= "11";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
