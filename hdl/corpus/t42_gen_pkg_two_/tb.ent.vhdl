-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two instances of one generic package
--!
--! Axis: generic. gp8 and gp4 both instantiated in the package file, one signal of each word_t.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : work.gp8.word_t := (others => '0');
    signal t : work.gp4.word_t := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= x"18";
        t <= x"5";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
