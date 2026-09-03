-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the child of the instance cost cases.
--!
--! Axis: the child with a signal driven by a process

library ieee;
    use ieee.std_logic_1164.all;

--! A child with one generic.
entity child is
    generic (k : integer := 0);
end entity;

architecture sim of child is
    signal c : std_ulogic := '0';
begin
    q: process
    begin
        wait for 10 ns;
        c <= '1';
        wait;
    end process;
end architecture;
