-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an in mode variable parameter
--!
--! Axis: a procedure with an in mode scalar variable parameter under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    procedure show(variable v : in integer; signal q : out std_ulogic) is
    begin
        q <= '1' when v > 0 else '0';
    end procedure;
begin
    p: process
        variable n : integer := 1;
    begin
        wait for 50 ns;
        show(n, s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
