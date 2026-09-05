-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a generate statement, with the type in the architecture
--!
--! Axis: protected method scopes. A shared variable of a protected type, its methods called from the only process, with a generate statement, with the type in the architecture after that process, to see which scope the second pair of method scopes hangs from.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type counter_t is protected
        procedure bump;
        impure function get return integer;
    end protected;

    type counter_t is protected body
        variable n : integer := 0;
        procedure bump is
        begin
            n := n + 1;
        end procedure;
        impure function get return integer is
        begin
            return n;
        end function;
    end protected body;

    shared variable ct : counter_t;
begin
    p: process
        variable v : integer := 0;
    begin
        wait for 50 ns;
        ct.bump;
        v := ct.get;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
    g: for i in 0 to 1 generate
        signal gs : std_ulogic := '0';
    begin
        gs <= s;
    end generate;
end architecture;
