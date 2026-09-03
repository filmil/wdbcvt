-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: one eight bit input port, left open, and one signal.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : in std_ulogic_vector(7 downto 0) := x"A5"
    );
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    s <= a(0);
end architecture;
