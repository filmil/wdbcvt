-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a field after an integer
--!
--! Axis: r.a <= '1' behind the integer field

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type ri_t is record
        i : integer;
        a : std_ulogic;
    end record;
    signal r : ri_t := (0, '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        r.a <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
