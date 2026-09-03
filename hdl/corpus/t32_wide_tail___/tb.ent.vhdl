-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a 300 byte field after a scalar
--!
--! Axis: r.v <= (others => '1') with r.a before it

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type wt_t is record
        a : std_ulogic;
        v : std_ulogic_vector(299 downto 0);
    end record;
    signal r : wt_t := ('0', (others => '0'));
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        r.v <= (others => '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
