-- SPDX-License-Identifier: Apache-2.0

--! @file
--! @brief Corpus case: a std_logic with 3 drivers, then a std_ulogic
--!
--! Axis: handles. A std_logic signal with 3 drivers declared before a second signal, to read the driver cost off the second handle.

library ieee;
    use ieee.std_logic_1164.all;

entity tb is
end entity;

architecture sim of tb is
    signal r : std_logic := 'Z';
    signal s : std_ulogic := '0';
begin
    q0: process
    begin
        wait for 6 ns;
        r <= 'Z';
        wait;
    end process;

    q1: process
    begin
        wait for 7 ns;
        r <= 'Z';
        wait;
    end process;

    p: process
    begin
        wait for 5 ns;
        r <= '1';
        s <= '1';
        wait for 10 ns;
        std.env.stop;
    end process;
end architecture;
