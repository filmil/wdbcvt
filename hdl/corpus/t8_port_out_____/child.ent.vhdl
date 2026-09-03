-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus child entity: one output port driven by a process.

library ieee;
    use ieee.std_logic_1164.all;

entity child is
    port (
        q : out std_ulogic
    );
end entity;

architecture sim of child is
begin
    p: process
    begin
        q <= '0';
        wait for 10 ns;
        q <= '1';
        wait;
    end process;
end architecture;
