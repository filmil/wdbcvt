-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a two dimensional array of vectors
--!
--! Axis: array (0 to 1, 0 to 2) of std_ulogic_vector(3 downto 0)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mat_t is array (0 to 1, 0 to 2) of std_ulogic_vector(3 downto 0);
    signal s : mat_t := (others => (others => x"0"));
begin
    p: process
    begin
        wait for 50 ns;
        s <= ((x"1", x"2", x"3"), (x"4", x"5", x"6"));
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
