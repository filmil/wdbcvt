-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one bit, assigned 1 twice in consecutive delta cycles at 10 ns.
--!
--! Axis: delta cycles. The second delta assigns the value the signal already has, to see whether a record needs a change of value.

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
        s <= '1';
        wait for 0 ns;
        s <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
