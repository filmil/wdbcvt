-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a procedure declared in a package
--!
--! Axis: a procedure of a package called from the process under -debug subprogram

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';

begin
    p: process
    begin
        wait for 50 ns;
        work.pk.drive(s, '1');
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
