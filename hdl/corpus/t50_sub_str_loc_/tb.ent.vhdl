-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a constrained string local
--!
--! Axis: a procedure with a local variable of a constrained string subtype under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure bump(signal q : out std_ulogic) is
        variable t : string(1 to 4) := "abcd";
    begin
        t(1) := 'x';
        q <= '1';
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
