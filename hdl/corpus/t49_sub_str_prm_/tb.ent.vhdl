-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an unconstrained string parameter
--!
--! Axis: a procedure with a constant string parameter of unconstrained size under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure drive(constant name : in string; signal q : out std_ulogic) is
    begin
        q <= '1';
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        drive("ab", s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
