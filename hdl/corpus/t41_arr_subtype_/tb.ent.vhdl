-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a constrained subtype of an unconstrained array type
--!
--! Axis: type. subtype byte_t is vec_t(3 downto -4), the signal declared with the subtype.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type vec_t is array (integer range <>) of std_ulogic;
    subtype byte_t is vec_t(3 downto -4);
    signal s : byte_t := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= x"18";
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
