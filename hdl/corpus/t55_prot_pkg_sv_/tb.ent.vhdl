-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package protected shared variable
--!
--! Axis: a shared variable of a protected type, both declared in the package, called through package subprograms under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';

begin
    p: process
        variable v : integer := 0;
    begin
        wait for 50 ns;
        work.pk.bump;
        v := work.pk.get;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
