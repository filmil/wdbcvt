-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a generic package with a constant of its generic
--!
--! Axis: generic. constant width : natural := n beside the subtype in gp, one instance gp8.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : work.gp8.word_t := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= x"18";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
