-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an access to an access type
--!
--! Axis: access and file entries. an access to an access type, to see whether the two words after the designated or element type move with that type.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_ptr is access integer;
    type ptr_ptr is access int_ptr;
begin
    p: process
        variable p : ptr_ptr;
    begin
        wait for 50 ns;
        p := new int_ptr'(new integer'(5));
        deallocate(p);
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
