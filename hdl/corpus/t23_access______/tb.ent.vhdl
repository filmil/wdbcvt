-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an access type variable
--!
--! Axis: type int_ptr is access integer; variable p : int_ptr in the process

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_ptr is access integer;
begin
    p: process
        variable p : int_ptr;
    begin
        wait for 50 ns;
        p := new integer'(5);
        s <= '1';
        deallocate(p);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
