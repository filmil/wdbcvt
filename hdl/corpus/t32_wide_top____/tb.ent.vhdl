-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the top 300 bytes of a 600 byte vector
--!
--! Axis: v(599 downto 300) <= (others => '1')

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(599 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        v(599 downto 300) <= (others => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
