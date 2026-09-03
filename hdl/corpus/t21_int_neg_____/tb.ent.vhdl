-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an integer signal taking a negative value
--!
--! Axis: s <= -165 in place of 165

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : integer := 0;
begin
    p: process
    begin
        wait for 50 ns;
        s <= -165;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
