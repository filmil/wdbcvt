-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record with two unconstrained fields
--!
--! Axis: type. alpha and bravo both unconstrained, the signal bundle_t(alpha(3 downto 0), bravo(7 downto 0)).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type bundle_t is record
        alpha : std_ulogic_vector;
        bravo : std_ulogic_vector;
    end record;
    signal s : bundle_t(alpha(3 downto 0), bravo(7 downto 0)) := (x"0", x"00");
begin
    p: process
    begin
        wait for 10 ns;
        s <= (x"5", x"A5");
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
