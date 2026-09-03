-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a string process variable.
--!
--! Axis: object kind. A string(1 to 5) process variable, to see whether it gets a declaration and a size the way an integer variable does.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
        variable v : string(1 to 5) := "hello";
    begin
        wait for 10 ns;
        v := "world";
        s <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
