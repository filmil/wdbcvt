-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief The innermost level of t2_hier3.

library ieee;
    use ieee.std_logic_1164.all;

entity leaf is
end entity;

architecture sim of leaf is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait;
    end process;
end architecture;
