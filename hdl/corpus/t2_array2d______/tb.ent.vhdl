-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of vectors with one transition.
--!
--! Axis: aggregate structure, the array case rather than the record
--! case. Four elements of eight bits each.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type row_array_t is array (0 to 3) of std_ulogic_vector(7 downto 0);
    signal s : row_array_t := (others => x"00");
begin
    p: process
    begin
        wait for 50 ns;
        s <= (x"A5", x"5A", x"0F", x"F0");
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
