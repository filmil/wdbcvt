-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two generics, the first of std_ulogic_vector(7 downto 0)
--!
--! Axis: a generic of std_ulogic_vector(7 downto 0) before an integer one, for the handle stride

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        a : std_ulogic_vector(7 downto 0) := x"00";
        b : integer := 1
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';

begin
    p: process
    begin
        wait for 50 ns;

        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
