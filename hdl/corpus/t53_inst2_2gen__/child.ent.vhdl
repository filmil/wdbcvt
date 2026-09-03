-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the child of the scope cost cases.
--!
--! Axis: a second generic of the child

library ieee;
    use ieee.std_logic_1164.all;

--! A child with one generic.
entity child is
    generic (k : integer := 0; k2 : integer := 7);
end entity;

architecture sim of child is

begin

end architecture;
