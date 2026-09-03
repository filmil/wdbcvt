-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two signals constraining one record differently
--!
--! Axis: type. s is bundle_t(bravo(7 downto 0)) and t is bundle_t(bravo(3 downto 0)).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type bundle_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector;
    end record;
    signal s : bundle_t(bravo(7 downto 0)) := ('0', x"00");
    signal t : bundle_t(bravo(3 downto 0)) := ('0', x"0");
begin
    p: process
    begin
        wait for 10 ns;
        s <= ('1', x"A5");
        t <= ('1', x"5");
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
