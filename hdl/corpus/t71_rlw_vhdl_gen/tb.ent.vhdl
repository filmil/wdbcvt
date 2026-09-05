-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a VHDL real generic
--!
--! Axis: the width of a real parameter. a VHDL real generic beside a logic, to see where the 16 bits a real parameter declares comes from, when a real variable declares 32 and both hold one float64.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        r : real := 1.5
    );
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
