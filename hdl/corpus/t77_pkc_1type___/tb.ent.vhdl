-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one type and no object
--!
--! Axis: package handle space. A package that declares one type and no object, read from the process, to see what each kind of declaration costs the handle space.

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
        a := integer(work.pk.small_t'(3));
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
