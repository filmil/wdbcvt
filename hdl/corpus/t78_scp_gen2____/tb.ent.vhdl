-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a generate of two iterations
--!
--! Axis: scope handle space. a generate of two iterations, to see what a scope costs beyond its objects, and what a generate iteration saves against an instance of the same body.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    g1: for i in 0 to 1 generate
        signal g : std_ulogic := '0';
    begin
        g <= s;
    end generate;

    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
