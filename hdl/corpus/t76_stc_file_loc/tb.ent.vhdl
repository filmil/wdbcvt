-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a file local of a procedure
--!
--! Axis: storage classes. a file local of a procedure, to see which storage class word 28 of the instance record gives it, and whether any form gives the 5 that no case has produced.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_file is file of integer;
    procedure w(v : integer) is
        file f : int_file open write_mode is "t76b.txt";
    begin
        write(f, v);
    end procedure;
begin
    p: process
        
    begin
        wait for 50 ns;
        w(5);
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
