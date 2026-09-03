-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: s'delayed
--!
--! Axis: d <= s'delayed(2 ns), an implicit signal behind the attribute

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal d : std_ulogic := '0';
begin
    d <= s'delayed(2 ns);
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
