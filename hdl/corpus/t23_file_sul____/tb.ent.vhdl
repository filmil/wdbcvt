-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a file of std_ulogic
--!
--! Axis: type sul_file is file of std_ulogic; file f : sul_file

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    type sul_file is file of std_ulogic;
    file f : sul_file;
begin
    p: process
    begin
        wait for 50 ns;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
