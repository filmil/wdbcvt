-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a nested record whose inner record holds an integer and a vector.
--!
--! Axis: nested record constraint. The scalar field is an integer, to see whether the extra triple is the scalar's own range.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type inner_t is record
        alpha : integer;
        bravo : std_ulogic_vector(3 downto 0);
    end record;
    type outer_t is record
        delta_f : inner_t;
        echo_f : std_ulogic;
    end record;
    signal s : outer_t := (delta_f => (alpha => 0, bravo => x"0"), echo_f => '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= (delta_f => (alpha => 165, bravo => x"A"), echo_f => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
