-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an unconstrained two dimensional array type
--!
--! Axis: array (natural range <>, natural range <>) of std_ulogic, constrained at the signal

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mat_t is array (natural range <>, natural range <>) of std_ulogic;
    signal s : mat_t(0 to 1, 0 to 2) := (others => (others => '0'));
begin
    p: process
    begin
        wait for 50 ns;
        s <= (('1', '0', '1'), ('0', '1', '1'));
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
