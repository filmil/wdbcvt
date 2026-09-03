-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: four transitions, 10 ns apart.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s <= '1';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        s <= '1';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
