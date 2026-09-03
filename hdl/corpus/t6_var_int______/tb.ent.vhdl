-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one bit and a process variable that changes twice.
--!
--! Axis: object kind. A process variable, to see whether its changes are logged or only its initial value.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
        variable v : integer := 7;
    begin
        wait for 10 ns;
        v := v + 1;
        s <= '1';
        wait for 10 ns;
        v := v + 1;
        s <= '0';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
