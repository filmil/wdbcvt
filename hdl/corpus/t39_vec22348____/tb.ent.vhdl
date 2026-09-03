-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one 22348 bit vector, one change.
--!
--! Axis: value size. 22348 bytes: the last chunk 296, exactly twice the chunk.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(22347 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
