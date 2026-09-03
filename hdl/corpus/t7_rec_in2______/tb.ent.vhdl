-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a nested record whose inner record holds two bytes.
--!
--! Axis: nested record constraint. The inner record is two scalars, 2 bytes, to see whether the extra triple's 8 tracks the inner size, its alignment or nothing.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type inner_t is record
        alpha : std_ulogic;
        bravo : std_ulogic;
    end record;
    type outer_t is record
        delta_f : inner_t;
        echo_f : std_ulogic;
    end record;
    signal s : outer_t := (delta_f => (alpha => '0', bravo => '0'), echo_f => '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= (delta_f => (alpha => '1', bravo => '1'), echo_f => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
