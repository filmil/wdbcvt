-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an in mode signal parameter
--!
--! Axis: a procedure with an in mode signal parameter and an out one under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure copy(signal a : in std_ulogic; signal q : out std_ulogic) is
    begin
        q <= not a;
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        copy(s, s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
