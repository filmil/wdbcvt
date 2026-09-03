-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a file local of a procedure
--!
--! Axis: a procedure with a local file object under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_file is file of integer;
    procedure drive(signal q : out std_ulogic) is
        file fl : int_file;
        variable v : integer := 0;
    begin
        v := v + 1;
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
