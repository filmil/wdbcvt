-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the child of the scope cost cases.
--!
--! Axis: the signal driven by a concurrent assignment

library ieee;
    use ieee.std_logic_1164.all;

--! A child with one generic.
entity child is
    generic (k : integer := 0);
end entity;

architecture sim of child is
    signal c : std_ulogic := '0';
begin
    c <= '1' after 10 ns;
end architecture;
