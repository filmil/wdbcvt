-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: one signal and a process variable.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
end entity;

architecture sim of child is
    signal s : std_ulogic := '0';
begin
    p: process
        variable v : integer := 0;
    begin
        wait for 10 ns;
        v := 3;
        s <= '1';
        wait;
    end process;
end architecture;
