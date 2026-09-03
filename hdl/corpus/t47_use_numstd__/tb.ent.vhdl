-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the tier 1 baseline with an unused use of numeric_std
--!
--! Axis: use clause. `use ieee.numeric_std.all` added to t1_bit_one_edge and nothing else, to see whether the 0x1f8 of handle space of the tier 2 cases is the package, not the type.

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.numeric_std.all;

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
