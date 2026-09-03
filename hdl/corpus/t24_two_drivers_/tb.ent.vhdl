-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two drivers of a std_logic
--!
--! Axis: processes p and q both drive r, resolved to X at 70 ns

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
        r <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;

    q: process
    begin
        r <= 'Z';
        wait for 70 ns;
        r <= '0';
        wait;
    end process;
end architecture;
