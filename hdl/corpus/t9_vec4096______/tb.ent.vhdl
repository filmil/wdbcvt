-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: two 4096 bit vectors, one change each.
--!
--! Axis: value size. A 4096 byte value is wider than the 0x800 arena span, to see how handles and arenas behave when one object is wider than a span.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(4095 downto 0) := (others => '0');
    signal t : std_ulogic_vector(4095 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => '1');
        wait for 10 ns;
        t <= (others => '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
