-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief The instantiated entity.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
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
