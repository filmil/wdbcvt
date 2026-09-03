-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two processes write one field of a resolved record
--!
--! Axis: r.b driven by p ('1' at 50 ns) and q ('Z' at 0, '0' at 70 ns), resolved to X

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type trio_t is record
        a : std_logic;
        b : std_logic;
        c : std_logic;
    end record;
    signal r : trio_t := ('0', 'Z', '0');
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        r.b <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
    q: process
    begin
        r.b <= 'Z';
        wait for 70 ns;
        r.b <= '0';
        wait;
    end process;
end architecture;
