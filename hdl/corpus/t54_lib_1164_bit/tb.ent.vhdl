-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a bit signal under std_logic_1164
--!
--! Axis: a bit signal under std_logic_1164

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : bit := '0';
begin
    p: process
        variable a : integer := 0;
    begin
        a := a + 1;
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
