-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an enumeration with the literals of bit
--!
--! Axis: type mybit_t is ('0', '1'), the literals of BIT under another name

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    type mybit_t is ('0', '1');
    signal s : mybit_t := '0';
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
