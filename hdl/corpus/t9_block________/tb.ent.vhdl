-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a block with its own signal.
--!
--! Axis: block statement. A labelled block declaring a signal, to see what unit kind it gets and how its signal is scoped.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal x : std_ulogic := '0';
begin
    b: block
        signal y : std_ulogic := '0';
    begin
        y <= x;
    end block;

    p: process
    begin
        wait for 10 ns;
        x <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
