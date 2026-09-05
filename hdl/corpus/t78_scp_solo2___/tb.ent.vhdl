-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two instances of a child with no ports
--!
--! Axis: scope handle space. two instances of a child with no ports, to see what a scope costs beyond its objects, and what a generate iteration saves against an instance of the same body.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    u: entity work.solo;
    v: entity work.solo;

    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
