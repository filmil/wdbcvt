-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the child of the scope cost cases.
--!
--! Axis: a variable in the child's process

library ieee;
    use ieee.std_logic_1164.all;

--! A child with one generic.
entity child is
    generic (k : integer := 0);
end entity;

architecture sim of child is

begin
    q: process
        variable v : integer := 0;
    begin
        v := v + 1;
        wait;
    end process;
end architecture;
