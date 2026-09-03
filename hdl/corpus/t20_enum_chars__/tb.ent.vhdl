-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an enumeration of three character literals
--!
--! Axis: type sym_t is ('a', 'b', 'c'), character literals that are not BIT's

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type sym_t is ('a', 'b', 'c');
    signal s : sym_t := 'a';
begin
    p: process
    begin
        wait for 50 ns;
        s <= 'c';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
