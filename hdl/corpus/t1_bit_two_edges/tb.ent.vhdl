-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: one bit with two transitions.
--!
--! Axis: transition count, 1 to 2. Isolates the size and shape of one
--! transition record.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 25 ns;
        s <= '0';
        wait for 25 ns;
        std.env.stop;
    end process;
end architecture;
