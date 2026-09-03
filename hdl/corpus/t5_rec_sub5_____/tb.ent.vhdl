-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a nested record whose inner record holds five scalars.
--!
--! Axis: nested record constraint. t2_record_nested has nine scalars inside the inner record; this one has five, to see which number the extra constraint triple tracks.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type inner_t is record
        alpha : std_ulogic;
        bravo : std_ulogic_vector(3 downto 0);
    end record;
    type outer_t is record
        delta_f : inner_t;
        echo_f : std_ulogic;
    end record;
    signal s : outer_t := (delta_f => (alpha => '0', bravo => x"0"), echo_f => '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= (delta_f => (alpha => '1', bravo => x"A"), echo_f => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
