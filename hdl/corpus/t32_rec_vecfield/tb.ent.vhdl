-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a slice of a record field
--!
--! Axis: r.v(3 downto 0) <= x"F"

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rv_t is record
        a : std_ulogic;
        v : std_ulogic_vector(7 downto 0);
    end record;
    signal r : rv_t := ('0', x"00");
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        r.v(3 downto 0) <= x"F";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
