-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a constrained subtype of a record
--!
--! Axis: type. subtype b8_t is bundle_t(bravo(7 downto 0)), the signal declared with the subtype.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type bundle_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector;
    end record;
    subtype b8_t is bundle_t(bravo(7 downto 0));
    signal s : b8_t := ('0', x"00");
begin
    p: process
    begin
        wait for 10 ns;
        s <= ('1', x"A5");
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
