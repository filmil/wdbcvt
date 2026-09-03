-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: log_wave issued at 0 and again at 10 ns.
--!
--! Axis: logging. A second log_wave -recursive * after run 10 ns, to see whether an already logged signal gets a second record or a second object.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 5 ns;
        s <= '1';
        wait for 10 ns;
        s <= '0';
        wait for 10 ns;
        s <= '1';
        wait for 5 ns;
        std.env.stop;
    end process;
end architecture;
