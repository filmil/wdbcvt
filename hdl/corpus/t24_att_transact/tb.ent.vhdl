-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: s'transaction
--!
--! Axis: t <= s'transaction, with a second assignment of the same value at 70 ns

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal t : bit := '0';
begin
    t <= s'transaction;
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 20 ns;
        s <= '1';
        wait for 30 ns;
        std.env.stop;
    end process;
end architecture;
