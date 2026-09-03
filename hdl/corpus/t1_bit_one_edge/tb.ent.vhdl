-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: one bit with one transition.
--!
--! The Tier 1 baseline. Every other Tier 1 case differs from this one
--! along exactly one axis.
--!
--! Differs from t0_bit_const by one transition, and by nothing else.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
