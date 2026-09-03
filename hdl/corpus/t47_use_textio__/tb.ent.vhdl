-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: the baseline with an unused use of textio
--!
--! Axis: use clause. `use std.textio.all`, a package with file objects and no constants.

library ieee;
    use ieee.std_logic_1164.all;
    use std.textio.all;

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
