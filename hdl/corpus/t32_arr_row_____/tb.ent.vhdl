-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a row of an array of vectors
--!
--! Axis: a(1) <= x"FF" of an array (0 to 3) of std_ulogic_vector(7 downto 0)

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type row_arr_t is array (0 to 3) of std_ulogic_vector(7 downto 0);
    signal a : row_arr_t := (others => x"00");
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        a(1) <= x"FF";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
