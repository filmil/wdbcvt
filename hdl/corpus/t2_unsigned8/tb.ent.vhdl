-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a `unsigned(7 downto 0)` signal with one transition.
--!
--! Axis: signal type. Differs from t1_bit_one_edge only in the type of
--! `s`, so the byte difference is what this type costs.

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.numeric_std.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : unsigned(7 downto 0) := x"00";
begin
    p: process
    begin
        wait for 50 ns;
        s <= x"A5";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
