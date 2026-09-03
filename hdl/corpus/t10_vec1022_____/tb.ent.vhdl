-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one 1022 bit vector, one change.
--!
--! Axis: value size. 1022 bytes, seven chunks of 146 exactly, to pin the chunk count rule.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(1021 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
