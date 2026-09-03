-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one 730 bit vector, one change.
--!
--! Axis: value size. 730 bytes is five chunks of 146, to see whether the count follows the size.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(729 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
