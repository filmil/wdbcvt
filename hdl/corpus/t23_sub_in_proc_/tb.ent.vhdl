-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a procedure local to a process
--!
--! Axis: the procedure declared in the process, as t9_proc_local, under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
        procedure flip is
            variable r : std_ulogic;
        begin
            r := not s;
            s <= r;
        end procedure;
    begin
        wait for 50 ns;
        flip;
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
