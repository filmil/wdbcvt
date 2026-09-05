-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a file of a forty element array
--!
--! Axis: access and file entries. a file of a forty element array, to see whether the two words after the designated or element type move with that type.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type arr_t is array (0 to 39) of integer;
    type arr_file is file of arr_t;
    file f : arr_file;
begin
    p: process
        
    begin
        wait for 50 ns;
        
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
