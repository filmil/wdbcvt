-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the baseline with a use of a package of one subtype
--!
--! Axis: use clause. `use work.typ_pkg.all`, a package with one subtype and no object, to price the scope and unit alone.

library ieee;
    use ieee.std_logic_1164.all;
    use work.typ_pkg.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
