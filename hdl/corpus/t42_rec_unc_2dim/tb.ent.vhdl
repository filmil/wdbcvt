-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a record with an unconstrained two dimensional field
--!
--! Axis: type. mat_t is array (natural range <>, natural range <>) of std_ulogic, the field m : mat_t, the signal rec_t(m(0 to 1, 0 to 2)).

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mat_t is array (natural range <>, natural range <>) of std_ulogic;
    type rec_t is record
        a : std_ulogic;
        m : mat_t;
    end record;
    signal s : rec_t(m(0 to 1, 0 to 2)) := ('0', (others => (others => '0')));
begin
    p: process
    begin
        wait for 10 ns;
        s <= ('1', (("101"), ("011")));
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
