-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a shared variable
--!
--! Axis: shared variable sv : integer := 0 in the architecture, assigned once

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
    shared variable sv : integer := 0;
begin
    p: process
    begin
        wait for 50 ns;
        sv := 5;
        s <= '1';
        wait for 50 ns;
        std.env.stop;
    end process;
end architecture;
