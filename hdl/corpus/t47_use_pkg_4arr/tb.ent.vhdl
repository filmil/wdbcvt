-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the baseline with a use of a package of four array types
--!
--! Axis: use clause. `use work.arr_pkg.all`, four unconstrained array types and no object, to see whether a library package's price is its types.

library ieee;
    use ieee.std_logic_1164.all;
    use work.arr_pkg.all;

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
