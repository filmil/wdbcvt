-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a file parameter
--!
--! Axis: a procedure with a file parameter under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_file is file of integer;
    procedure put(file f : int_file; signal q : out std_ulogic) is
    begin
        write(f, 1);
        q <= '1';
    end procedure;
    file fo : int_file open write_mode is "t51.bin";
begin
    p: process
    begin
        wait for 50 ns;
        put(fo, s);
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
