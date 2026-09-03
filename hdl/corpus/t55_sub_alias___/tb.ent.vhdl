-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an alias of a local variable
--!
--! Axis: a procedure with an alias of a local variable under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure drive(signal q : out std_ulogic) is
        variable v : integer := 0;
        alias b : integer is v;
    begin
        b := 1;
        q <= '1';
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        drive(s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
