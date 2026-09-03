-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a process vector variable with a literal initialiser
--!
--! Axis: a process variable of std_ulogic_vector(3 downto 0) initialised from a literal, before an integer one under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';

begin
    p: process
        variable a : std_ulogic_vector(3 downto 0) := "0000";
        variable b : integer := 0;
    begin
        wait for 50 ns;
        a(0) := '1';
        b := b + 1;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
