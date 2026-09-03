-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a vector constant parameter
--!
--! Axis: a procedure with a constrained vector constant parameter under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure drive(constant a : in std_ulogic_vector(3 downto 0); signal q : out std_ulogic) is
    begin
        q <= a(0);
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        drive("0001", s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
