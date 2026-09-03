-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: a scalar out port.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        a : out std_ulogic := '0'
    );
end entity;

architecture sim of child is
begin
    q: process
    begin
        wait for 50 ns;
        a <= '1';
        wait;
    end process;
end architecture;
