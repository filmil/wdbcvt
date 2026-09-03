-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a plain package with only a subtype
--!
--! Axis: package. subtype word_t is std_ulogic_vector(7 downto 0) in a package without generics, against the generic package instance.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : work.pp.word_t := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= x"18";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
