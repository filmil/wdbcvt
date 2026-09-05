-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a file of a record
--!
--! Axis: access and file entries. a file of a record, to see whether the two words after the designated or element type move with that type.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type rec_t is record
        a : std_ulogic;
        n : integer;
    end record;
    type rec_file is file of rec_t;
    file f : rec_file;
begin
    p: process
        
    begin
        wait for 50 ns;
        
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
