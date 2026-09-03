-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one transition at 1000 ns rather than 10 ns. Same count, later time.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 1000 ns;
        s <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
