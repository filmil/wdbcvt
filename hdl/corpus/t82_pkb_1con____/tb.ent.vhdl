-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a package whose body declares one integer constant
--!
--! Axis: a package body in the handle space. a package whose body declares one integer constant, to see whether what a package body declares lands in the package's block or past the second region.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        k : integer := 7
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
        variable a : integer := 0;
    begin
        wait for 50 ns;
        a := k + work.pk.f;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
