-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one 22349 bit vector, one change.
--!
--! Axis: value size. 22349 bytes: the last chunk 297, one over twice the chunk.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(22348 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
