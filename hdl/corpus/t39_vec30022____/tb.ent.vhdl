-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one 30022 bit vector, one change.
--!
--! Axis: value size. 30022 bytes: 202 chunks of 148, the last 274 bytes, one record.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(30021 downto 0) := (others => '0');
begin
    p: process
    begin
        wait for 10 ns;
        s <= (others => '1');
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
