-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: three input ports, left open, and one signal.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : in std_ulogic := '1';
        b : in std_ulogic := '0';
        c : in std_ulogic := '1'
    );
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    s <= c;
end architecture;
