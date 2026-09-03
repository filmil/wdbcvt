-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the run split into three commands.
--!
--! Axis: logging. run 10 ns, run 10 ns, run -all instead of one run -all, to see whether the database notices the breaks.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 5 ns;
        s <= '1';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        s <= '1';
        wait for 5 ns;
        std.env.stop;
    end process;
end architecture;
