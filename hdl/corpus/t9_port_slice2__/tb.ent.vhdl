-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a two bit input port bound to the middle of a four bit vector.
--!
--! Axis: port association. The actual is a slice in the middle of a vector, to see whether the instance record carries a byte offset into the actual.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic_vector(3 downto 0) := "0000";
begin
    dut: entity work.child port map (a => x(2 downto 1));

    p: process
    begin
        wait for 10 ns;
        x <= "0110";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
