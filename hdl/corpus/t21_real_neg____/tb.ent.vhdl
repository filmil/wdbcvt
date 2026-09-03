-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a real signal taking a negative value
--!
--! Axis: s <= -1.5 in place of 1.5

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : real := 0.0;
begin
    p: process
    begin
        wait for 50 ns;
        s <= -1.5;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
