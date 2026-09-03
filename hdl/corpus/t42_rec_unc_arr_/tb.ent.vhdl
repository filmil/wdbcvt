-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array of records with an unconstrained field
--!
--! Axis: type. arr_t is array (0 to 1) of bundle_t, the signal arr_t(open)(bravo(3 downto 0)).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type bundle_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector;
    end record;
    type arr_t is array (0 to 1) of bundle_t;
    signal s : arr_t(open)(bravo(3 downto 0)) := (others => ('0', x"0"));
begin
    p: process
    begin
        wait for 10 ns;
        s <= (('1', x"5"), ('0', x"A"));
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
