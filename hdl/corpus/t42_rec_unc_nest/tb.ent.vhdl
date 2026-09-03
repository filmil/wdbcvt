-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record whose inner record has an unconstrained field
--!
--! Axis: type. inner_t has v : std_ulogic_vector, outer_t holds i : inner_t and b : std_ulogic, the signal outer_t(i(v(3 downto 0))).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type inner_t is record
        v : std_ulogic_vector;
    end record;
    type outer_t is record
        i : inner_t;
        b : std_ulogic;
    end record;
    signal s : outer_t(i(v(3 downto 0))) := ((v => x"0"), '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= ((v => x"5"), '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
