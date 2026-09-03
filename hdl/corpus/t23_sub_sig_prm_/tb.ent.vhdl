-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a signal parameter
--!
--! Axis: procedure drive(signal q : out std_ulogic; constant v : in std_ulogic) under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure drive(signal q : out std_ulogic; constant v : in std_ulogic) is
    begin
        q <= v;
    end procedure;
begin
    p: process
    begin
        wait for 50 ns;
        drive(s, '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
