-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief The instantiated entity for the generics cases.
--!
--! The generic deliberately does not affect the signal, its type or its
--! transition, so any difference between cases is the generic binding
--! and nothing else.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    generic (
        --! Bound differently per case. Unused by the body on purpose.
        k : natural := 4
    );
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        wait;
    end process;
end architecture;
