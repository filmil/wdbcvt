-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: an array variable of the process, not of the function
--!
--! Axis: a array type declared in the architecture, a process variable of it beside the function under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type arr_t is array (0 to 3) of integer;
    function f(c : std_ulogic) return std_ulogic is
        variable v : integer := 0;
    begin
        v := v + 1;
        return c;
    end function;
begin
    p: process
        variable pa : arr_t := (others => 0);
    begin
        wait for 50 ns;
        s <= f('1');
        pa(1) := 1;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
