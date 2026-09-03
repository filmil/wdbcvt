-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record with a constrained and an unconstrained field
--!
--! Axis: type. alpha is std_ulogic_vector(3 downto 0) and bravo unconstrained, the signal bundle_t(bravo(7 downto 0)).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type bundle_t is record
        alpha : std_ulogic_vector(3 downto 0);
        bravo : std_ulogic_vector;
    end record;
    signal s : bundle_t(bravo(7 downto 0)) := (x"0", x"00");
begin
    p: process
    begin
        wait for 10 ns;
        s <= (x"5", x"A5");
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
