-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two top level entities, both logged.
--!
--! Axis: elaboration. The two tops of t45_two_tops with log_wave -recursive /tb2 and /tb, since -recursive * and /* reach the current scope /tb2 only, to see how the objects of two roots are keyed and paged.

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
