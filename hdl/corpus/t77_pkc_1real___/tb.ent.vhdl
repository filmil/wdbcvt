-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one real constant
--!
--! Axis: package handle space. A package that declares one real constant, read from the process, to see what each kind of declaration costs the handle space.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
        variable a : integer := 0;
    begin
        a := integer(work.pk.r);
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
