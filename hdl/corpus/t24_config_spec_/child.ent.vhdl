-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief The instantiated entity with two architectures.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
end entity;

architecture a of child is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        wait;
    end process;
end architecture;

architecture b of child is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 20 ns;
        s <= '1';
        wait;
    end process;
end architecture;
