-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: the child instance of t1_hier1.
--!
--! Holds the signal that t1_bit_one_edge holds at the top level, so
--! that the only difference between the two cases is one level of
--! hierarchy.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait;
    end process;
end architecture;
