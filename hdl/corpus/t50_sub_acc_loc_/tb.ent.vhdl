-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an access type local
--!
--! Axis: a procedure with a local variable of an access type under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_ptr is access integer;
    procedure bump(signal q : out std_ulogic) is
        variable p : int_ptr;
    begin
        p := new integer'(1);
        q <= '1';
        deallocate(p);
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        bump(s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
