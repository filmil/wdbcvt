-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record whose field is itself a record.
--!
--! Axis: aggregate nesting. Says whether nesting is represented, and
--! whether the inner type gets its own entry in the type table.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type inner_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector(7 downto 0);
    end record;
    type outer_t is record
        delta_f : inner_t;
        echo_f : std_ulogic;
    end record;
    signal s : outer_t := (delta_f => (alpha => '0', bravo => x"00"), echo_f => '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= (delta_f => (alpha => '1', bravo => x"A5"), echo_f => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
