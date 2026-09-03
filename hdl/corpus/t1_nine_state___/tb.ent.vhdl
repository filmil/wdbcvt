-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: one bit taking all nine std_ulogic values.
--!
--! Axis: the value alphabet. Gives the code for each of U, X, 0, 1, Z,
--! W, L, H and the dash, and says how many bits one value costs.
--!
--! The signal starts at 'U' by taking the type default, so the first
--! logged value is 'U' and the transitions walk the rest in order.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic;
begin
    p: process
    begin
        wait for 10 ns;
        s <= 'X';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        s <= '1';
        wait for 10 ns;
        s <= 'Z';
        wait for 10 ns;
        s <= 'W';
        wait for 10 ns;
        s <= 'L';
        wait for 10 ns;
        s <= 'H';
        wait for 10 ns;
        s <= '-';
        wait for 20 ns;
        std.env.stop;
    end process;
end architecture;
