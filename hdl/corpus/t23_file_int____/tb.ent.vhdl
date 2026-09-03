-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a file of integer
--!
--! Axis: type int_file is file of integer; file f : int_file

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type int_file is file of integer;
    file f : int_file;
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
