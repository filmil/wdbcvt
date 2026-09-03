-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer field assigned
--!
--! Axis: r.i <= 5 with r.a untouched

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
        r.i <= 5;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
