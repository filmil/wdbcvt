-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one bit, assigned the value it already holds at 10 ns.
--!
--! Axis: no change. One assignment of the current value, to see whether an assignment without a change of value leaves a record.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
