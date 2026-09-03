-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a slice of an out port bound to a slice
--!
--! Axis: child port v : out std_ulogic_vector(3 downto 0), port map (v => x(3 downto 0)), v(1 downto 0) <= "11" inside the child

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal x : std_ulogic_vector(7 downto 0) := x"00";
begin
    dut: entity work.child port map (v => x(3 downto 0));
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
