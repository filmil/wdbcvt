-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: one 8-bit vector with one transition.
--!
--! Axis: width, 1 to 8. The name stays one character long, so only the
--! width moves. The value x"A5" is asymmetric on purpose, so that bit
--! order is visible in the encoding.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic_vector(7 downto 0) := x"00";
begin
    p: process
    begin
        wait for 50 ns;
        s <= x"A5";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
