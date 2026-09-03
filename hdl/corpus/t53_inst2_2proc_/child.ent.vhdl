-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the child of the scope cost cases.
--!
--! Axis: two processes in the child

library ieee;
    use ieee.std_logic_1164.all;

--! A child with one generic.
entity child is
    generic (k : integer := 0);
end entity;

architecture sim of child is

begin
    q: process
    begin
        wait;
    end process;
    r: process
    begin
        wait;
    end process;
end architecture;
