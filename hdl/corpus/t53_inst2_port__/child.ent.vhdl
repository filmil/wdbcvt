-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the child of the scope cost cases.
--!
--! Axis: an input port of the child connected to s

library ieee;
    use ieee.std_logic_1164.all;

--! A child with one generic.
entity child is
    generic (k : integer := 0);
    port (a : in std_ulogic);
end entity;

architecture sim of child is

begin

end architecture;
