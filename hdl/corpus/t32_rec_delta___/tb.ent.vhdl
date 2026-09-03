-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two fields in two deltas
--!
--! Axis: r.a <= '1'; wait for 0 ns; r.b <= '1'

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type trio_t is record
        a : std_ulogic;
        b : std_ulogic;
        c : std_ulogic;
    end record;
    signal r : trio_t := ('0', '0', '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        r.a <= '1';
        wait for 0 ns;
        r.b <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
