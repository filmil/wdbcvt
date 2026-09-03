-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a variable parameter
--!
--! Axis: a procedure with an inout variable parameter under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure bump(variable v : inout integer) is
    begin
        v := v + 1;
    end procedure;
begin
    p: process
        variable n : integer := 0;
    begin
        wait for 50 ns;
        bump(n);
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
