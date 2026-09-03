-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a 300 byte field at the start of a record
--!
--! Axis: r.v <= (others => '1') with r.a behind it

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type wf_t is record
        v : std_ulogic_vector(299 downto 0);
        a : std_ulogic;
    end record;
    signal r : wf_t := ((others => '0'), '0');
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
