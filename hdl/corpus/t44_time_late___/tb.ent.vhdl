-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an edge at 1 ns and one at 5 ms.
--!
--! Axis: time width. A first change at 1 ns and a second at 5 ms, so the page's second word, the span, is the one past 2^32.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 1 ns;
        s <= '1';
        wait for 5 ms;
        s <= '0';
        wait for 1 ns;
        std.env.stop;
    end process;
end architecture;
