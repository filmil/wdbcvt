-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a slice of a to vector
--!
--! Axis: v(0 to 3) <= x"F" of an 8 bit to vector

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    signal v : std_ulogic_vector(0 to 7) := x"00";
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        v(0 to 3) <= x"F";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
