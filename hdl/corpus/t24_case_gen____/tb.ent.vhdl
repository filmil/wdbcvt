-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: case generate
--!
--! Axis: g: case k generate, constant k := 1, the second alternative taken

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    constant k : integer := 1;
begin
    g: case k generate
        when zero: 0 =>
            signal a : std_ulogic := '0';
        begin
            a <= '1' after 50 ns;
        end;
        when one: others =>
            signal b : std_ulogic := '0';
        begin
            b <= '1' after 50 ns;
        end;
    end generate;

    p: process
    begin
        wait for 100 ns;
        std.env.stop;
    end process;
end architecture;
