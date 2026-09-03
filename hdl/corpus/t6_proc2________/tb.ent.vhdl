-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two processes, each with a variable and a signal.
--!
--! Axis: scope count. Two processes, to see how process scopes and their variables are numbered.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal a : std_ulogic := '0';
    signal b : std_ulogic := '0';
begin
    p: process
        variable v : integer := 1;
    begin
        wait for 10 ns;
        v := v + 1;
        a <= '1';
        wait;
    end process;
    q: process
        variable w : integer := 2;
    begin
        wait for 20 ns;
        w := w + 1;
        b <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
