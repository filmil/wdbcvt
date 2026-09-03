-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the baseline with a use of an empty package
--!
--! Axis: use clause. `use work.emp_pkg.all`, a package with nothing in it, to price the scope and unit alone.

library ieee;
    use ieee.std_logic_1164.all;
    use work.emp_pkg.all;

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
