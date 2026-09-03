-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: one transition to 'Z' rather than '1'. Same time, different value.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal s : std_ulogic := '0';
begin
    p: process
    begin
        wait for 10 ns;
        s <= 'Z';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
