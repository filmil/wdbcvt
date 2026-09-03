-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a bit signal under the usual std_logic_1164 use clause
--!
--! Axis: use clause. The `bit` signal of t47_use_none with `library ieee` and `use ieee.std_logic_1164.all` back, to separate the type from the use clause.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : bit := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
