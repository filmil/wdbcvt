-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: one bit whose name is 40 characters long.
--!
--! Axis: name length, 1 to 40. Reveals whether names sit inline or in a
--! table, and whether they are length prefixed or terminated.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal ssssssssssssssssssssssssssssssssssssssss : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        ssssssssssssssssssssssssssssssssssssssss <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
