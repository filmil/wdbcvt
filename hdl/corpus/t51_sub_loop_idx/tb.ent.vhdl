-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a loop index inside a procedure
--!
--! Axis: a procedure with a for loop under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure drive(signal q : out std_ulogic) is
        variable n : integer := 0;
    begin
        for i in 0 to 3 loop
            n := n + i;
        end loop;
        q <= '1' when n = 6 else '0';
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
