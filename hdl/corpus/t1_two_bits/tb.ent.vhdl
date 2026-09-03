-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: two bits, one transition each.
--!
--! Axis: signal count, 1 to 2. Both names are one character long, so
--! name length stays constant and only the count moves.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal t : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        t <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
