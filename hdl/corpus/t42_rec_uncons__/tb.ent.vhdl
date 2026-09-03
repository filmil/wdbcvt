-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record with an unconstrained vector field
--!
--! Axis: type. VHDL-2008 record whose field bravo is an unconstrained std_ulogic_vector, the signal constrained as bundle_t(bravo(7 downto 0)).

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
begin
    p: process
    begin
        wait for 10 ns;
        s <= ('1', x"A5");
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
