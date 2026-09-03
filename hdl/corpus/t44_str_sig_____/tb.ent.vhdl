-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a string signal.
--!
--! Axis: type: string. A string(1 to 5) signal, to see how a VHDL string is typed, sized and recorded.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : string(1 to 5) := "hello";
begin
    p: process
    begin
        wait for 50 ns;
        s <= "world";
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
