-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a slice driven concurrently
--!
--! Axis: v(3 downto 0) <= (others => s) as a concurrent assignment

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(7 downto 0) := x"00";
begin
    v(3 downto 0) <= (others => s);
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
