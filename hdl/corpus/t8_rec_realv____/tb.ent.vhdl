-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a nested record whose inner record holds a real and a vector.
--!
--! Axis: nested record constraint. A real beside a vector, to see what range a real field contributes.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type inner_t is record
        alpha : real;
        bravo : std_ulogic_vector(3 downto 0);
    end record;
    type outer_t is record
        delta_f : inner_t;
        echo_f : std_ulogic;
    end record;
    signal s : outer_t := (delta_f => (alpha => 0.0, bravo => x"0"), echo_f => '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= (delta_f => (alpha => 1.5, bravo => x"A"), echo_f => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
