-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a nested record whose inner record starts with the vector.
--!
--! Axis: nested record constraint. The vector field first, to see whether the extra triple's numbers track the field position.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type inner_t is record
        bravo : std_ulogic_vector(3 downto 0);
        alpha : std_ulogic;
    end record;
    type outer_t is record
        delta_f : inner_t;
        echo_f : std_ulogic;
    end record;
    signal s : outer_t := (delta_f => (bravo => x"0", alpha => '0'), echo_f => '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= (delta_f => (bravo => x"A", alpha => '1'), echo_f => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
