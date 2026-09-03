-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: three edges under the default script.
--!
--! Axis: logging. The bench of the tier under the default script, so that the late and partial logs have a baseline with the same edges.

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
