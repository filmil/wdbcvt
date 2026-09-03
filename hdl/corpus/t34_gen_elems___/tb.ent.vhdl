-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: elements driven from a for generate
--!
--! Axis: g: for i in 0 to 2 generate v(i) <= s; three concurrent drivers of one vector

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(3 downto 0) := "0000";
begin
    g: for i in 0 to 2 generate
        v(i) <= s;
    end generate;
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
