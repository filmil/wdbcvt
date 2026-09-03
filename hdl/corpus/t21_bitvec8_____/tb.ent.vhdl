-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a bit_vector signal
--!
--! Axis: bit_vector(7 downto 0) in place of std_ulogic_vector

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : bit_vector(7 downto 0) := x"00";
begin
    p: process
    begin
        wait for 50 ns;
        s <= x"a5";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
