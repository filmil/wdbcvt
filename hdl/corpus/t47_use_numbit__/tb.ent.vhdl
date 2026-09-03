-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the baseline with an unused use of numeric_bit
--!
--! Axis: use clause. `use ieee.numeric_bit.all`, a package with the same two null constants as numeric_std.

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.numeric_bit.all;

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
