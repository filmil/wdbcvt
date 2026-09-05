-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: no package
--!
--! Axis: where a package sits in the handle space. no package, read from the process, to see whether the package moves the generic and the process variable that come after the signals.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
    generic (
        k : integer := 7
    );
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
        variable a : integer := 0;
    begin
        wait for 50 ns;
        a := k;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
