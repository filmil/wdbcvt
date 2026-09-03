-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a `time` signal with one transition.
--!
--! Axis: signal type. Differs from t1_bit_one_edge only in the type of
--! `s`, so the byte difference is what this type costs.

library ieee;
    use ieee.std_logic_1164.all;
    use ieee.numeric_std.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : time := 0 ns;
begin
    p: process
    begin
        wait for 50 ns;
        s <= 7 ns;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
