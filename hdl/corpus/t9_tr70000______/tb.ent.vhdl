-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one bit, 70000 transitions 1 ns apart.
--!
--! Axis: page count. 70000 one byte records need 117 pages, more than the 100 slots of an arena record, to see where the directory goes when it overflows.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        for i in 1 to 70000 loop
            wait for 1 ns;
            s <= not s;
        end loop;
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
