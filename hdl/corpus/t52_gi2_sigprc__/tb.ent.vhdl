-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a two iteration generate, the iteration holding a signal driven by a process
--!
--! Axis: the iteration holding a signal driven by a process

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    g: for i in 0 to 1 generate
        signal c : std_ulogic := '0';
    begin
        q: process
        begin
            wait for 10 ns;
            c <= '1';
            wait;
        end process;
    end generate;

    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
