-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one driver of a std_logic assigning at time 0
--!
--! Axis: q assigns r <= 'Z', the initial value, at time 0, and p assigns '1' at 50 ns from the same process

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal r : std_logic := 'Z';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
    q: process
    begin
        r <= 'Z';
        wait for 50 ns;
        r <= '1';
        wait;
    end process;
end architecture;
