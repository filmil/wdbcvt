-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: 70000 generate iterations with a signal each
--!
--! Axis: scale. A for generate of 70000 iterations each declaring a signal, to see whether a scope count or index above 65535 is stored whole.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is

begin
    g: for i in 0 to 69999 generate
        signal s : std_ulogic := '0';
    begin
        q: process
        begin
            if i = 0 then
                wait for 5 ns;
                s <= '1';
            elsif i = 69999 then
                wait for 15 ns;
                s <= '1';
            end if;
            wait;
        end process;
    end generate;

    p: process
    begin
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;
