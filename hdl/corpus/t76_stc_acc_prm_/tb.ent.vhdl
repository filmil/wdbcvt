-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an access parameter of a procedure
--!
--! Axis: storage classes. an access parameter of a procedure, to see which storage class word 28 of the instance record gives it, and whether any form gives the 5 that no case has produced.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_ptr is access integer;
    procedure w(p : inout int_ptr) is
    begin
        p := new integer'(5);
    end procedure;
begin
    p: process
        variable q : int_ptr;
    begin
        wait for 50 ns;
        w(q);
        deallocate(q);
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
