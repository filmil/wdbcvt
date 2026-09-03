-- SPDX-License-Identifier: Apache-2.0
--! @file
--! @brief Corpus case: one bit that never changes.
--!
--! Differs from t0_nosig by one signal, and by nothing else. The byte
--! difference is one signal record with no transitions.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 100 ns;
        std.env.stop;
    end process;
end architecture;
